// Package geheimtipp serves the geheimtipp site's own backend under this app's
// origin.
//
// geheimtipp is a separate, running service with its own database — the football
// pool at ihleven.de. Its code is not ours to change, so its API cannot learn to
// accept cross-origin calls from wherever this app happens to be served. The
// browser has to reach it through here instead.
//
// The upstream today is not that service directly but the geheimtipp frontend's
// own server layer, which forwards to it. That layer goes away with the frontend
// it belongs to, so this must
// eventually point at the Go service itself — and whether this proxy survives at
// all then depends on whether the site is reconfigured to route those paths
// there. Either way the upstream is configuration, and the frontend reads its
// base from configuration too, so neither answer is baked in.
package geheimtipp

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Identifier says who is making a request, and whether they may act as
// themselves on the pool.
//
// Declared here, where it is used, rather than taken from the auth package:
// this needs a name and a yes, and depending on the whole account type for that
// would be a cycle waiting to happen.
type Identifier interface {
	// PoolLogin returns the login to speak as, and false for a request that
	// should be forwarded anonymously.
	PoolLogin(r *http.Request) (string, bool)
}

// Proxy forwards everything under its mount point to upstream.
//
// It is deliberately not gated on an account. What it forwards to is already
// readable by anyone — the editions, the matchdays, the rankings — and the pages
// that use it are public for the same reason.
//
// What it does add is identity. Writing a tip is protected by the upstream's
// own token, and that token is now minted here from this app's session rather
// than obtained by a second sign-in: whoever is signed in here is who the pool
// is told about. A request with no session carries no token — not even one the
// browser offers — and the pool answers it as it answers any visitor, so the
// public pages stay public.
//
// Both id and minter may be nil — a deployment with no pool secret configured
// still proxies, anonymously.
func Proxy(upstream string, prefix string, id Identifier, minter *Minter) (func(http.ResponseWriter, *http.Request) error, error) {
	target, err := url.Parse(strings.TrimSuffix(upstream, "/"))
	if err != nil {
		return nil, fmt.Errorf("geheimtipp: upstream %q is not a URL: %w", upstream, err)
	}
	if target.Scheme == "" || target.Host == "" {
		return nil, fmt.Errorf("geheimtipp: upstream %q needs a scheme and a host", upstream)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			// Upstream is a different site, and its nginx routes by Host. Sending
			// ours would land somewhere else, or nowhere.
			r.Out.Host = target.Host

			// SetURL joins the inbound path onto the target's. What arrives here
			// carries our mount point, which upstream knows nothing about.
			r.Out.URL.Path = singleJoin(target.Path, strings.TrimPrefix(r.In.URL.Path, prefix))

			// The pool's credential is this app's to issue, so a token the
			// browser sent is never forwarded — not even when nothing replaces
			// it. See dropToken.
			dropToken(r.Out)

			// Speak as whoever is signed in here. r.In is the request as the
			// browser sent it, which is where the session cookie is; r.Out is
			// what upstream will see, which is where the pool's token goes.
			if id == nil || minter == nil {
				return
			}
			if login, ok := id.PoolLogin(r.In); ok {
				if err := minter.authorize(r.Out, login); err != nil {
					// Forwarding anonymously is the honest failure: the pool
					// answers as it does for a visitor, rather than the request
					// failing in a way that reads as the pool being down.
					slog.Error("geheimtipp: minting a token", "login", login, "err", err)
				}
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			scopeCookies(resp.Header, prefix)
			return nil
		},
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		proxy.ServeHTTP(w, r)
		return nil
	}, nil
}

// scopeCookies rewrites every Set-Cookie to belong to this mount point and to
// this host.
//
// Domain goes because a cookie set for the upstream's domain would be dropped by
// the browser, which sees this response as coming from us; without the attribute
// it binds to whichever host we are served from, which is what makes signing in
// work from here at all.
//
// Path is forced to the mount point because the upstream sets its session cookie
// for "/" — correct on its own site, wrong here, where "/" is the family app.
// The two sites share an origin and nothing else, and an unscoped credential is
// where that would blur first.
func scopeCookies(h http.Header, prefix string) {
	cookies := h.Values("Set-Cookie")
	if len(cookies) == 0 {
		return
	}

	rewritten := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		parts := strings.Split(cookie, ";")
		kept := parts[:0]
		for _, part := range parts {
			switch attr := strings.ToLower(strings.TrimSpace(part)); {
			case strings.HasPrefix(attr, "domain="), strings.HasPrefix(attr, "path="):
				continue
			}
			kept = append(kept, part)
		}
		rewritten = append(rewritten, strings.Join(kept, ";")+"; Path="+prefix)
	}
	h.Del("Set-Cookie")
	for _, cookie := range rewritten {
		h.Add("Set-Cookie", cookie)
	}
}

// singleJoin joins two path segments with exactly one slash between them.
func singleJoin(base, rest string) string {
	switch {
	case base == "" || base == "/":
		return "/" + strings.TrimPrefix(rest, "/")
	case rest == "" || rest == "/":
		return base
	}
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(rest, "/")
}
