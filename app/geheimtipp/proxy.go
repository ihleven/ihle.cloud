// Package geheimtipp serves the geheimtipp site's own backend under this app's
// origin.
//
// geheimtipp is a separate, running service with its own database — the football
// pool at ihleven.de. Its code is not ours to change, so its API cannot learn to
// accept cross-origin calls from wherever this app happens to be served. The
// browser has to reach it through here instead.
//
// The upstream today is not that service directly but the geheimtipp frontend's
// own server layer, which forwards to it and mints the session cookie on the way
// back. That layer goes away with the frontend it belongs to, so this must
// eventually point at the Go service itself — and whether this proxy survives at
// all then depends on whether the site is reconfigured to route those paths
// there. Either way the upstream is configuration, and the frontend reads its
// base from configuration too, so neither answer is baked in.
package geheimtipp

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/interhome-group/cms/pkg/errs"
)

// Proxy forwards everything under its mount point to upstream.
//
// It is deliberately not gated on an account. What it forwards to is already
// readable by anyone — the editions, the matchdays, the rankings — and the pages
// that use it are public for the same reason. Writing a tip is protected by the
// upstream's own token, which this passes along without inspecting.
func Proxy(upstream string, prefix string) (func(http.ResponseWriter, *http.Request) error, error) {
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
		},
		ModifyResponse: func(resp *http.Response) error {
			// A cookie set for the upstream's domain would be dropped by the
			// browser, which sees this response as coming from us. Without the
			// attribute it binds to whichever host we are served from, which is
			// what makes signing in work from here at all.
			stripCookieDomain(resp.Header)
			return nil
		},
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		proxy.ServeHTTP(w, r)
		return nil
	}, nil
}

// stripCookieDomain removes the Domain attribute from every Set-Cookie, leaving
// the rest of each cookie untouched.
func stripCookieDomain(h http.Header) {
	cookies := h.Values("Set-Cookie")
	if len(cookies) == 0 {
		return
	}

	rewritten := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		parts := strings.Split(cookie, ";")
		kept := parts[:0]
		for _, part := range parts {
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(part)), "domain=") {
				continue
			}
			kept = append(kept, part)
		}
		rewritten = append(rewritten, strings.Join(kept, ";"))
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

// Unconfigured reports that no upstream was given, per request rather than at
// startup: the rest of the app runs perfectly well without the pool, and a route
// that says so is easier to diagnose than one that is silently absent.
func Unconfigured(w http.ResponseWriter, _ *http.Request) error {
	return errs.New("the geheimtipp backend is not configured",
		errs.HTTPStatus(http.StatusNotImplemented))
}
