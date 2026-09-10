package common

import "testing"

func TestParseFriendCodeRoundTrip(t *testing.T) {
	for _, game := range []string{"RMCP", "HDMP", "ADAE"} { // md5, reversed crc8, plain crc8
		const pid = 123456789
		shown := CalcFriendCodeString(pid, game)
		raw, reversed, ok := ParseFriendCode(shown)
		if !ok {
			t.Fatalf("%s: %q did not parse", game, shown)
		}
		if uint32(raw) != pid && uint32(reversed) != pid {
			t.Errorf("%s: %q gives pids %d/%d, want %d", game, shown, uint32(raw), uint32(reversed), pid)
		}
		if !FriendCodeMatches(shown, pid, game) {
			t.Errorf("%s: %q does not verify against its own profile", game, shown)
		}
		if FriendCodeMatches(shown, pid+1, game) {
			t.Errorf("%s: %q verifies against a different profile", game, shown)
		}
	}
	if _, _, ok := ParseFriendCode("1234-5678"); ok {
		t.Error("an 8-digit code parsed")
	}
}
