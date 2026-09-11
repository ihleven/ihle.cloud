package authn

import (
	"net/url"
	"strings"
)

// EnrollPath is where an enrollment link lands: a server route, distinct from
// the SPA's /enroll page it redirects to. If the two shared a path the server
// route would shadow the page and it could never render.
const EnrollPath = "/auth/enroll"

// EnrollURL builds the link handed to someone enrolling a passkey.
//
// The token travels in the query string only as far as that route, which moves
// it into a cookie and redirects — so it does not linger in the address bar, the
// history or a Referer header.
func EnrollURL(base, token string) string {
	return strings.TrimSuffix(base, "/") + EnrollPath + "?t=" + url.QueryEscape(token)
}
