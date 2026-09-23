package gpcm

// OpenPak bans (website/docs/ban-lookup.md). WFC has no login of its own: a Wii or DS is an
// OpenPak account only once a player links its console user id on the website, and the core
// holds that link under the "wfc" namespace. So the website's internal ban lookup is asked
// by (namespace "wfc", subject = console user id):
//
//   - at GPCM login, where WiiLink already refuses its own banned profiles, a banned account is
//     answered the same way (error 22002, "You are banned ... Terms of Service");
//   - every 30 s for every logged-in session, and a session whose account is now banned is
//     kicked with the "You have been banned" message (22002), the path WiiLink's own ban uses.
//
// A console nobody has linked is not an OpenPak account and keeps playing. When the lookup
// cannot be answered a login is let through (the next sweep catches a banned one): refusing
// would lock every unlinked Wii and DS out whenever the website is down.
//
// Configured by WEBSITE_INTERNAL_URL and WEBSITE_INTERNAL_KEY; with either unset the gate is off.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"wwfc/logging"
)

const (
	banActiveTTL  = 30 * time.Second // an "allowed" answer is trusted this long; "banned" never is
	banSweepEvery = 30 * time.Second
)

// banVerdict is the gate's answer about one console.
type banVerdict struct {
	Banned bool
	Reason string // shown to the player when set
}

type banGate struct {
	url, key string
	client   *http.Client
	now      func() time.Time

	mu      sync.Mutex
	allowed map[uint64]time.Time
}

func newBanGate(baseURL, key string) *banGate {
	if baseURL == "" || key == "" {
		return nil
	}
	return &banGate{
		url: strings.TrimRight(baseURL, "/"), key: key,
		client: &http.Client{Timeout: 5 * time.Second},
		now:    time.Now, allowed: map[uint64]time.Time{},
	}
}

var bans = newBanGate(os.Getenv("WEBSITE_INTERNAL_URL"), os.Getenv("WEBSITE_INTERNAL_KEY"))

// check asks about one console user id. An error means the lookup could not be answered.
func (b *banGate) check(userID uint64) (banVerdict, error) {
	if b == nil || userID == 0 {
		return banVerdict{}, nil
	}
	b.mu.Lock()
	if exp, ok := b.allowed[userID]; ok && b.now().Before(exp) {
		b.mu.Unlock()
		return banVerdict{}, nil
	}
	b.mu.Unlock()

	q := url.Values{"namespace": {"wfc"}, "subject": {strconv.FormatUint(userID, 10)}}
	req, err := http.NewRequest(http.MethodGet, b.url+"/internal/bans?"+q.Encode(), nil)
	if err != nil {
		return banVerdict{}, err
	}
	req.Header.Set("Authorization", "Bearer "+b.key)
	resp, err := b.client.Do(req)
	if err != nil {
		return banVerdict{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))

	var answer struct {
		Banned bool   `json:"banned"`
		Reason string `json:"reason"`
		Error  string `json:"error"`
	}
	_ = json.Unmarshal(body, &answer)

	var v banVerdict
	switch {
	case resp.StatusCode == http.StatusOK:
		v = banVerdict{Banned: answer.Banned, Reason: answer.Reason}
	case resp.StatusCode == http.StatusNotFound && strings.Contains(answer.Error, "no account for that subject"):
		// Not linked: not an OpenPak account, nothing to ban.
	case resp.StatusCode == http.StatusNotFound:
		v = banVerdict{Banned: true} // linked to an account that no longer exists: locked out
	default:
		return banVerdict{}, fmt.Errorf("ban lookup: %s", resp.Status)
	}

	b.mu.Lock()
	if v.Banned {
		delete(b.allowed, userID)
	} else {
		b.allowed[userID] = b.now().Add(banActiveTTL)
	}
	b.mu.Unlock()
	return v, nil
}

// refuseIfBanned answers a GPCM login for a banned account with WiiLink's ban message and
// reports whether it did.
func (g *GameSpySession) refuseIfBanned(userID uint64) bool {
	v, err := bans.check(userID)
	if err != nil {
		logging.Warn(g.ModuleName, "OpenPak ban lookup failed, letting the login through:", err)
		return false
	}
	if !v.Banned {
		return false
	}
	g.replyError(GPError{
		ErrorCode:   ErrLogin.ErrorCode,
		ErrorString: "The OpenPak account linked to this console is banned.",
		Fatal:       true,
		WWFCMessage: WWFCMsgProfileBannedTOS,
		Reason:      v.Reason,
	})
	return true
}

// watchBans kicks logged-in sessions whose account became banned. Only a positive answer
// kicks: a lookup outage leaves everyone connected.
func watchBans() {
	if bans == nil {
		return
	}
	for {
		time.Sleep(banSweepEvery)
		sweepBans()
	}
}

func sweepBans() {
	type online struct {
		profileID uint32
		userID    uint64
	}
	var players []online
	mutex.Lock()
	for profileID, s := range sessions {
		if s.LoggedIn && s.User.UserId != 0 {
			players = append(players, online{profileID, s.User.UserId})
		}
	}
	mutex.Unlock()

	for _, p := range players {
		v, err := bans.check(p.userID)
		if err != nil {
			logging.Warn("GPCM", "OpenPak ban sweep:", err)
			return // the website is not answering: try the whole sweep again next time
		}
		if v.Banned {
			logging.Notice("GPCM", "OpenPak account banned, kicking profile", p.profileID)
			KickPlayer(p.profileID, "banned")
		}
	}
}
