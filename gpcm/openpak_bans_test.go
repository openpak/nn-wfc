package gpcm

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBanGateCheck(t *testing.T) {
	answers := map[string]struct {
		status int
		body   string
	}{
		"1": {200, `{"account_id":"a","banned":true,"status":"banned","reason":"cheating"}`},
		"2": {200, `{"account_id":"b","banned":false,"status":"active"}`},
		"3": {404, `{"error":"no account for that subject"}`},
		"4": {404, `{"error":"account not found"}`},
		"5": {502, `{"error":"core"}`},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/bans" || r.Header.Get("Authorization") != "Bearer k" || r.URL.Query().Get("namespace") != "wfc" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		a := answers[r.URL.Query().Get("subject")]
		w.WriteHeader(a.status)
		w.Write([]byte(a.body))
	}))
	defer srv.Close()
	g := newBanGate(srv.URL, "k")

	cases := []struct {
		user       uint64
		banned     bool
		wantErr    bool
		wantReason string
	}{
		{1, true, false, "cheating"},
		{2, false, false, ""},
		{3, false, false, ""}, // not linked: not an OpenPak account
		{4, true, false, ""},  // linked to an account that is gone
		{5, false, true, ""},  // lookup down
	}
	for _, c := range cases {
		v, err := g.check(c.user)
		if (err != nil) != c.wantErr || v.Banned != c.banned || v.Reason != c.wantReason {
			t.Errorf("user %d: got %+v, %v", c.user, v, err)
		}
	}
	if _, ok := g.allowed[1]; ok {
		t.Error("a banned answer must not be cached")
	}
	if _, ok := g.allowed[2]; !ok {
		t.Error("an allowed answer is cached")
	}
	if newBanGate("", "k") != nil || newBanGate(srv.URL, "") != nil {
		t.Error("the gate is off unless both URL and key are set")
	}
	var off *banGate
	if v, err := off.check(1); v.Banned || err != nil {
		t.Error("an off gate bans nobody")
	}
}
