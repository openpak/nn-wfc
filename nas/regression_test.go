// Regression checks for the C10 online-play surface (prds/platform-wii-ds-prd.md
// WD-3): the wire shapes every Wii and DS title depends on before a single
// packet of gameplay flows — the NAS auth form, the reply encoding, DLS1 DLC
// counts, the connection test and the host routing that Traefik's port-80
// routers feed. Database-free by design: these pin the protocol, and the
// protocol is what a regression would break.
package nas

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wwfc/common"
)

// The server reads game_list.tsv from its working directory (sake's package
// init needs it too), so the tests run from the repository root exactly where
// the deployed process runs.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// --- the NAS auth form and its reply ---

func dwcForm(t *testing.T, values map[string]string) string {
	t.Helper()
	form := url.Values{}
	for key, value := range values {
		form.Set(key, common.Base64DwcEncoding.EncodeToString([]byte(value)))
	}
	return form.Encode()
}

func authRequest(t *testing.T, form string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://nas.nintendowifi.net/ac", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", "nas.nintendowifi.net")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	return req
}

func TestAuthFormDecodesDSValues(t *testing.T) {
	// unitcd 0 = DS: little-endian UTF-16, as the DS sends it.
	ingamesn := common.UTF16Encode("MARIO", binary.LittleEndian)

	req := authRequest(t, dwcForm(t, map[string]string{
		"unitcd":   "0",
		"gamecd":   "ADAE",
		"ingamesn": string(ingamesn),
	}))
	fields, err := parseAuthRequest(req)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if string(fields["gamecd"]) != "ADAE" {
		t.Fatalf("gamecd = %q", fields["gamecd"])
	}
	if got := common.UTF16Decode(fields["ingamesn"], binary.LittleEndian); got != "MARIO" {
		t.Fatalf("ingamesn = %q, want MARIO", got)
	}
}

func TestAuthFormDecodesWiiStringsBigEndian(t *testing.T) {
	// unitcd 1 = Wii: big-endian UTF-16. A Wii name read little-endian is
	// garbage — this is the regression that would show up as mojibake.
	be := common.UTF16Encode("LINK", binary.BigEndian)

	req := authRequest(t, dwcForm(t, map[string]string{
		"unitcd":   "1",
		"ingamesn": string(be),
	}))
	fields, err := parseAuthRequest(req)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := common.UTF16Decode(fields["ingamesn"], binary.BigEndian); got != "LINK" {
		t.Fatalf("ingamesn = %q, want LINK", got)
	}
}

func TestAuthFormRejectsUndecodableValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://nas.nintendowifi.net/ac",
		strings.NewReader("gamecd=!!!not-base64***"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()
	if _, err := parseAuthRequest(req); err == nil {
		t.Fatal("an undecodable form value must fail the parse, not pass through")
	}
}

func TestAuthReplyEncodingRoundTrips(t *testing.T) {
	rec := httptest.NewRecorder()
	writeAuthResponse(rec, map[string]string{
		"returncd": "100",
		"token":    "\xff\xfe\x01binary-ish*value",
	})

	body := rec.Body.String()
	if !strings.HasSuffix(body, "\x00") {
		t.Fatal("DWC reads the reply as a NUL terminated string; the NUL is missing")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("content type = %q", ct)
	}
	if n := rec.Header().Get("Content-Length"); n == "" {
		t.Fatal("no content length: DWC sizes the buffer before reading")
	}

	params, err := url.ParseQuery(strings.TrimSuffix(body, "\x00"))
	if err != nil {
		t.Fatalf("reply is not a query string: %v", err)
	}
	if got, err := common.Base64DwcEncoding.DecodeString(params.Get("returncd")); err != nil || string(got) != "100" {
		t.Fatalf("returncd round trip = %q, %v", got, err)
	}
	if got, err := common.Base64DwcEncoding.DecodeString(params.Get("token")); err != nil || string(got) != "\xff\xfe\x01binary-ish*value" {
		t.Fatalf("token round trip = %q, %v", got, err)
	}
}

// --- DLS1 (DLC) ---

func dlsRequest(t *testing.T, host, action, rhgamecd string) *http.Request {
	t.Helper()
	form := map[string]string{"action": action}
	if rhgamecd != "" {
		form["rhgamecd"] = rhgamecd
	}
	req := httptest.NewRequest(http.MethodPost, "http://"+host+"/download",
		strings.NewReader(dwcForm(t, form)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", host)
	return req
}

func TestDLS1CountsDLCFiles(t *testing.T) {
	oldDir := dlcDir
	t.Cleanup(func() { dlcDir = oldDir })
	dlcDir = t.TempDir()

	for game, files := range map[string]int{"RMCE": 2, "ABCD": 0} {
		dir := filepath.Join(dlcDir, game)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < files; i++ {
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("dlc%d.bin", i)), []byte{1}, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	rec := httptest.NewRecorder()
	handleDownloadEndpoint(rec, dlsRequest(t, "dls1.nintendowifi.net", "count", "RMCE"))
	if rec.Code != http.StatusOK || rec.Body.String() != "2\x00" {
		t.Fatalf("count = %d %q, want 200 \"2\\x00\"", rec.Code, rec.Body.String())
	}

	// A game with no DLC on the server still gets an answer, never an error:
	// the console treats a failed count as failed DLC.
	rec = httptest.NewRecorder()
	handleDownloadEndpoint(rec, dlsRequest(t, "dls1.nintendowifi.net", "count", "NOPE"))
	if rec.Code != http.StatusOK || rec.Body.String() != "0\x00" {
		t.Fatalf("missing game = %d %q, want 200 \"0\\x00\"", rec.Code, rec.Body.String())
	}
}

// TestDLS1RejectsBadGameCodes pins the input validation: the game code lands
// in a filesystem path, so anything non-alphanumeric must be refused.
func TestDLS1RejectsBadGameCodes(t *testing.T) {
	for _, bad := range []string{"rmce", "RMC", "RMCE1", "RMC-E"} {
		rec := httptest.NewRecorder()
		handleDownloadEndpoint(rec, dlsRequest(t, "dls1.nintendowifi.net", "count", bad))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("rhgamecd %q: code = %d, want 400", bad, rec.Code)
		}
	}
}

// --- host routing ---

func TestHostRouting(t *testing.T) {
	cases := []struct {
		host string
		want *http.ServeMux
	}{
		{"nas.nintendowifi.net", authMux},
		{"naswii.nintendowifi.net", authMux},
		{"dls1.nintendowifi.net", dlsMux},
		{"sake.gs.nintendowifi.net", sakeMux},
		{"gamestats.gs.nintendowifi.net", gamestatsMux},
		{"gamestats2.gamespy.com", gamestatsMux},
		{"race.gs.nintendowifi.net", raceMux},
		{"nas.openpak.org", authMux},
		{"dls1.openpak.org", dlsMux},
	}
	for _, c := range cases {
		if got := muxForHost(c.host); got != c.want {
			t.Fatalf("host %q routed to the wrong handler", c.host)
		}
	}
	if muxForHost("conntest.nintendowifi.net") != nil {
		t.Fatal("conntest must fall through to the default mux")
	}
}

// --- the reachability pages a console checks before it believes the network ---

func TestConnectionTestPage(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://conntest.nintendowifi.net/", nil)
	handleConnectionTest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "This is test.html page") {
		t.Fatal("the exact page the console looks for is missing")
	}
	if rec.Header().Get("X-Organization") != "Nintendo" {
		t.Fatal("X-Organization header missing")
	}
}

func TestNASTestEndpoint(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://nas.nintendowifi.net/nastest.jsp", nil)
	handleNASTest(rec, req)

	body, _ := io.ReadAll(rec.Body)
	if !bytes.Contains(body, []byte("AuthServer is up")) {
		t.Fatal("nastest.jsp no longer reports the server as up")
	}
	if rec.Header().Get("Server") != "Nintendo" {
		t.Fatal("Server header missing")
	}
}
