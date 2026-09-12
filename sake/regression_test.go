// Regression checks for the SAKE half of C10 (prds/platform-wii-ds-prd.md
// WD-3): the SOAP request surface every friend-race/leaderboard storage call
// arrives through. These pin the identity checks that run before the
// database — unknown game, wrong secret key — and the rule that a malformed
// request is answered, not fatal. The record operations behind a *valid*
// identity have their own database path and are covered upstream of here.
package sake

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// The server reads game_list.tsv from its working directory, so the tests run
// from the repository root exactly where the deployed process runs.
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

const envelope = `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <CreateRecord xmlns="http://gamespy.net/sake">
      <gameid>%d</gameid>
      <secretKey>%s</secretKey>
      <loginTicket>proposal.1.100.11000.0</loginTicket>
      <tableid>MostWins</tableid>
    </CreateRecord>
  </soap:Body>
</soap:Envelope>`

func storageRequest(t *testing.T, action, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost,
		"http://sake.gs.nintendowifi.net/SakeStorageServer/StorageServer.asmx", strings.NewReader(body))
	req.Host = "sake.gs.nintendowifi.net"
	req.Header.Set("SOAPAction", action)
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	rec := httptest.NewRecorder()
	handleStorageRequest(rec, req)
	return rec
}

// A missing SOAPAction never reaches a handler and never panics; the console
// retries rather than hard-failing on an empty answer.
func TestMissingSOAPActionAnswersEmpty(t *testing.T) {
	rec := storageRequest(t, "", "not even xml")
	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Fatalf("missing SOAPAction: %d %q, want an empty 200", rec.Code, rec.Body.String())
	}
}

func TestMalformedXMLEnvelopeAnswersEmpty(t *testing.T) {
	rec := storageRequest(t, "http://gamespy.net/sake/CreateRecord", "not even xml")
	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Fatalf("malformed XML: %d %q, want an empty 200", rec.Code, rec.Body.String())
	}
}

// An unknown game id stops at the identity check: answered with a SAKE result,
// not a database round trip and not a crash.
func TestUnknownGameIDIsRejected(t *testing.T) {
	rec := storageRequest(t, "http://gamespy.net/sake/CreateRecord",
		"<Envelope xmlns=\"http://schemas.xmlsoap.org/soap/envelope/\"><Body><CreateRecord xmlns=\"http://gamespy.net/sake\"><gameid>999999</gameid></CreateRecord></Body></Envelope>")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "DatabaseUnavailable") {
		t.Fatalf("unknown game answered %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/xml" {
		t.Fatalf("content type = %q, want text/xml", ct)
	}
}

// A known game with the wrong secret key is the standard forged-client case:
// refused at the identity check, before anything per-profile happens.
func TestWrongSecretKeyIsRejected(t *testing.T) {
	// mariokartwii is game 1687 with secret 9r3Rmy; send a wrong one.
	rec := storageRequest(t, "http://gamespy.net/sake/CreateRecord",
		"<Envelope xmlns=\"http://schemas.xmlsoap.org/soap/envelope/\"><Body><CreateRecord xmlns=\"http://gamespy.net/sake\"><gameid>1687</gameid><secretKey>AAAAAA</secretKey></CreateRecord></Body></Envelope>")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "SecretKeyInvalid") {
		t.Fatalf("wrong secret answered %q", rec.Body.String())
	}
}

// The header action and the envelope's element must name the same operation —
// the mismatch is answered with a well-formed but empty SOAP envelope, never
// dispatched to an operation and never left unanswered.
func TestMismatchedSOAPActionAndBodyAnswerEmpty(t *testing.T) {
	body := "<Envelope xmlns=\"http://schemas.xmlsoap.org/soap/envelope/\"><Body><CreateRecord xmlns=\"http://gamespy.net/sake\"><gameid>1687</gameid><secretKey>9r3Rmy</secretKey></CreateRecord></Body></Envelope>"
	rec := storageRequest(t, "http://gamespy.net/sake/UpdateRecord", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<s:Body></s:Body>") {
		t.Fatalf("mismatched action answered %q, want an empty SOAP body", rec.Body.String())
	}
}
