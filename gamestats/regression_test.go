// Regression checks for the gamestats half of C10 (prds/platform-wii-ds-prd.md
// WD-3): the HTTP leaderboard endpoint a WFC title polls during a race.
// Database-free: the token contract, the request hash check and the response
// framing are what a regression would break; the PD storage behind the UDP
// side has its own path.
package gamestats

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"wwfc/common"
)

// The server reads game_list.tsv from its working directory, so the tests run
// from the repository root exactly where the deployed process runs.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func webRequest(t *testing.T, target string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Host = "gamestats.gs.nintendowifi.net"
	return req
}

// A leaderboard that is not in the game list is a 404, never a panic and
// never an answer some other game would have accepted.
func TestUnknownGameIs404(t *testing.T) {
	rec := httptest.NewRecorder()
	HandleWebRequest(rec, webRequest(t, "http://gamestats.gs.nintendowifi.net/nosuchgamehere/web/client/get2.asp?pid=1"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", rec.Code)
	}
}

// The token a console requests before it can sign its leaderboard query.
func TestTokenIsStableAndPerPID(t *testing.T) {
	webSalt = "regression-salt"
	serverName = "test"

	first := httptest.NewRecorder()
	HandleWebRequest(first, webRequest(t, "http://gamestats.gs.nintendowifi.net/acrossingds/?pid=42"))
	second := httptest.NewRecorder()
	HandleWebRequest(second, webRequest(t, "http://gamestats.gs.nintendowifi.net/acrossingds/?pid=42"))
	other := httptest.NewRecorder()
	HandleWebRequest(other, webRequest(t, "http://gamestats.gs.nintendowifi.net/acrossingds/?pid=43"))

	if first.Code != http.StatusOK {
		t.Fatalf("code = %d", first.Code)
	}
	token := first.Body.String()
	if len(token) != 32 {
		t.Fatalf("token = %q (%d chars), want 32", token, len(token))
	}
	if token != second.Body.String() {
		t.Fatal("the same request produced two different tokens; the console signs against the first")
	}
	if token == other.Body.String() {
		t.Fatal("different pids produced the same token")
	}
}

// get2.asp with a correctly signed query: the framed RNK_GET payload. This is
// the exact exchange every leaderboard-capable title performs mid-game.
func TestGet2AnswersSignedQuery(t *testing.T) {
	webSalt = "regression-salt"
	serverName = "test"

	for name, versioned := range map[string]bool{
		"acrossingds":   false, // GameStatsVersion 1: payload + padding, no trailing hash
		"smashbrosxwii": true,  // GameStatsVersion 3: sha1 appended
	} {
		game := common.GetGameInfoByName(name)
		if game == nil {
			t.Fatalf("game list has no %s", name)
		}

		target := &url.URL{Scheme: "http", Host: "gamestats.gs.nintendowifi.net",
			Path: "/" + name + "/web/client/get2.asp", RawQuery: "pid=42"}
		token := calculateToken(target, target.Host)
		hasher := sha1.New()
		hasher.Write([]byte(game.GameStatsKey))
		hasher.Write([]byte(token))
		signed := target.String() + "&hash=" + hex.EncodeToString(hasher.Sum(nil))

		rec := httptest.NewRecorder()
		HandleWebRequest(rec, webRequest(t, signed))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: code = %d", name, rec.Code)
		}
		body := rec.Body.Bytes()

		// RNK_GET frame: version 1, count 0, then DWC's 13 bytes of padding.
		const payloadLen = 8 + 13
		if len(body) < payloadLen {
			t.Fatalf("%s: body = %d bytes, want at least %d", name, len(body), payloadLen)
		}
		if v := binary.LittleEndian.Uint32(body[0:4]); v != 1 {
			t.Fatalf("%s: RNK_GET version = %d", name, v)
		}
		if count := binary.LittleEndian.Uint32(body[4:8]); count != 0 {
			t.Fatalf("%s: RNK_GET count = %d, want 0", name, count)
		}

		if !versioned {
			if len(body) != payloadLen {
				t.Fatalf("%s: version 1 must not append a hash, got %d bytes", name, len(body))
			}
			continue
		}

		// Version > 1 signs the whole response: sha1(key + b64(payload) + key).
		want := game.GameStatsKey + base64.URLEncoding.EncodeToString(body[:payloadLen]) + game.GameStatsKey
		sum := sha1.Sum([]byte(want))
		if !strings.HasSuffix(string(body[payloadLen:]), hex.EncodeToString(sum[:])) {
			t.Fatalf("%s: trailing signature does not match", name)
		}
	}
}

// A wrong hash is tolerated with a warning — DWC pads and signs the answer
// regardless, so an unsigned client still gets a framed response.
func TestGet2ToleratesBadHash(t *testing.T) {
	webSalt = "regression-salt"
	serverName = "test"

	rec := httptest.NewRecorder()
	HandleWebRequest(rec, webRequest(t,
		"http://gamestats.gs.nintendowifi.net/acrossingds/web/client/get2.asp?pid=42&hash=deadbeef"))
	if rec.Code != http.StatusOK || len(rec.Body.Bytes()) < 8+13 {
		t.Fatalf("bad hash answered %d with %d bytes; want a framed 200", rec.Code, len(rec.Body.Bytes()))
	}
}
