package passkey

import "testing"

func TestCheckOrigin(t *testing.T) {
	tests := []struct {
		origin, rpID string
		ok           bool
	}{
		{"https://ihle.cloud", "ihle.cloud", true},
		{"https://www.ihle.cloud", "ihle.cloud", true},
		{"http://localhost:8000", "localhost", true},
		{"http://127.0.0.1:8000", "127.0.0.1", true},
		// A subdomain of localhost always resolves to this machine, and is how
		// two apps on one dev box avoid sharing a relying-party id.
		{"http://ihlvn.localhost:8000", "ihlvn.localhost", true},
		{"http://app.ihlvn.localhost:8000", "ihlvn.localhost", true},
		// Still not a licence for any plain-http host.
		{"http://notlocalhost.example", "notlocalhost.example", false},
		// A suffix match must not be a substring match.
		{"https://evilihle.cloud", "ihle.cloud", false},
		{"http://ihle.cloud", "ihle.cloud", false}, // not https
		{"https://example.com", "ihle.cloud", false},
		{"not a url at all", "ihle.cloud", false},
		{"https://", "ihle.cloud", false},
	}
	for _, tt := range tests {
		err := checkOrigin(tt.origin, tt.rpID)
		if tt.ok && err != nil {
			t.Errorf("checkOrigin(%q, %q) = %v, want ok", tt.origin, tt.rpID, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("checkOrigin(%q, %q) was accepted", tt.origin, tt.rpID)
		}
	}
}

// Without a session or an enrollment link, anyone could attach a passkey to any
// account.
