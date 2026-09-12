package hiauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"userid"`
	Alias        string `json:"alias"`
	Scope        string `json:"scope"`
}

func NewTokenMngr(store TokenStore) *TokenMngr {
	mngr := TokenMngr{
		ctx:        context.Background(),
		refreshers: map[string]*TokenRefresher{},
		authclient: authclient{http.DefaultClient, os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET")},
		TokenStore: store,
	}
	tokens, err := store.LoadTokens()
	if err != nil {
		log.Fatal("could not load tokens:", err.Error())
	}
	// Only the refresh token is stored; the access token is obtained on start.
	for _, m := range tokens {
		token := Token{
			TokenType:    "Bearer",
			RefreshToken: m["refresh_token"],
			Alias:        m["alias"],
			Scope:        m["scope"],
		}
		mngr.refreshers[token.Alias] = mngr.newTokenRefresher(token)
	}
	return &mngr
}

// NewTokenChecker returns a manager that can exchange a refresh token but keeps
// no tokens and starts no refresh loops — for asking HiDrive whether a stored
// credential still works, without taking over refreshing it.
func NewTokenChecker(clientID, clientSecret string) *TokenMngr {
	return &TokenMngr{
		ctx:        context.Background(),
		refreshers: map[string]*TokenRefresher{},
		authclient: authclient{http.DefaultClient, clientID, clientSecret},
	}
}

type authclient struct {
	httpclient *http.Client
	// BaseURL      string `arg:"env:OAUTH_BASE_URL" `
	ClientID     string `arg:"env:OAUTH_CLIENT_ID" `
	ClientSecret string `arg:"env:OAUTH_CLIENT_SECRET" `
	// RedirectURI  string `arg:"env:OAUTH_REDIRECT_URI" `
	// Resource     string `arg:"env:OAUTH_RESOURCE" `
	// Scope        string `arg:"env:OAUTH_SCOPE" `
}

type TokenMngr struct {
	ctx        context.Context
	refreshers map[string]*TokenRefresher
	// tokenURL overrides the HiDrive token endpoint; empty means the real one.
	tokenURL string
	authclient
	TokenStore
}

func (s *TokenMngr) GetTokenRefresher(key string) *TokenRefresher {
	refr, ok := s.refreshers[key]
	if ok {
		return refr
	}
	return nil
}

const (
	// Refresh this long before the access token expires, so one handed out now
	// outlives the request it is used for.
	lifeSpanSafetyMargin = 1 * time.Minute

	// Never schedule sooner than this. The wait is expiry minus the safety
	// margin, which is negative whenever the token is short-lived or already
	// expired — and a negative duration makes time.After fire at once, turning
	// the loop into a request flood.
	minRefreshDelay = 5 * time.Second

	// A failed refresh backs off from minRetryDelay, doubling to maxRetryDelay.
	minRetryDelay = 5 * time.Second
	maxRetryDelay = 15 * time.Minute

	tokenEndpoint = "https://my.hidrive.com/oauth2/token"
)

// OAuthError is the error the token endpoint reported, kept structured so the
// refresh loop can tell "try again later" from "this will never work".
type OAuthError struct {
	Code string
	Desc string
}

func (e *OAuthError) Error() string { return e.Code + ": " + e.Desc }

// Permanent reports whether retrying is pointless. A rejected refresh token
// stays rejected until a human redoes the authorization flow, so continuing to
// ask only hammers the provider.
func (e *OAuthError) Permanent() bool {
	return e.Code == "invalid_grant" || e.Code == "invalid_client"
}

// merged applies a refresh response to the token it refreshed.
//
// The response carries a new access token and little else: HiDrive does not
// return the refresh token, and may omit alias and scope. Replacing the token
// wholesale therefore blanks the very credential the next refresh needs — so
// each field is kept unless the response actually supplies one.
func (t Token) merged(resp Token) Token {
	t.AccessToken = resp.AccessToken
	t.ExpiresIn = resp.ExpiresIn

	if resp.RefreshToken != "" {
		t.RefreshToken = resp.RefreshToken
	}
	if resp.Alias != "" {
		t.Alias = resp.Alias
	}
	if resp.Scope != "" {
		t.Scope = resp.Scope
	}
	if resp.UserID != "" {
		t.UserID = resp.UserID
	}
	if resp.TokenType != "" {
		t.TokenType = resp.TokenType
	}

	return t
}

// newTokenRefresher starts the goroutine that keeps one alias's access token
// current and hands it to callers over a channel.
//
// Access tokens are not stored, so there is never one at startup: the first
// refresh happens immediately, and until it finishes callers block rather than
// receive an empty token.
func (s *TokenMngr) newTokenRefresher(token Token) *TokenRefresher {

	a := &TokenRefresher{
		stream: make(chan tokenResponse),
		token:  token,
	}

	go func() {
		var (
			err       error
			retry     = minRetryDelay
			haveToken bool
			giveUp    bool
		)

		refreshNow := time.NewTimer(0)
		defer refreshNow.Stop()

		for {
			// Offer the token only once there is one to offer, or once refreshing
			// has been abandoned and the error is the answer. A nil channel never
			// becomes ready, so a caller waits instead of getting nothing.
			var out chan tokenResponse
			if haveToken || giveUp {
				out = a.stream
			}

			select {
			case out <- tokenResponse{token: a.token, Err: err}:

			case <-refreshNow.C:
				t, e := s.RefreshToken(a.token.RefreshToken)
				if e != nil {
					err = e

					var oauthErr *OAuthError
					if errors.As(e, &oauthErr) && oauthErr.Permanent() {
						// Asking again cannot help; stop until the process is
						// restarted with a token someone has re-authorized.
						log.Printf("hi: refresh for %q abandoned: %v", a.token.Alias, e)
						giveUp = true

						continue
					}

					log.Printf("hi: refresh for %q failed, retrying in %s: %v", a.token.Alias, retry, e)
					refreshNow.Reset(retry)

					if retry *= 2; retry > maxRetryDelay {
						retry = maxRetryDelay
					}

					continue
				}

				a.token = a.token.merged(*t)
				haveToken = true
				err = nil
				retry = minRetryDelay

				refreshNow.Reset(refreshDelay(time.Duration(t.ExpiresIn) * time.Second))

			case <-s.ctx.Done():
				return
			}
		}
	}()

	return a
}

// refreshDelay is how long to wait before renewing a token that lives for
// expiresIn, never so short that renewing becomes a loop.
func refreshDelay(expiresIn time.Duration) time.Duration {
	if d := expiresIn - lifeSpanSafetyMargin; d > minRefreshDelay {
		return d
	}

	return minRefreshDelay
}

type TokenRefresher struct {
	stream chan tokenResponse
	token  Token
}

type tokenResponse struct {
	token Token
	Err   error
}

func (a *TokenRefresher) GetAccessToken() (string, error) {
	if a == nil {
		return "", errors.New("nil refresher")
	}
	tokenresponse := <-a.stream
	// 	return t.token.AccessToken, t.Err
	// workaround
	// fmt.Println("GetAccessToken", tokenresponse.token.Alias, tokenresponse.token.AccessToken)
	return tokenresponse.token.AccessToken, tokenresponse.Err
}

func (m *TokenMngr) RefreshToken(refreshtoken string) (*Token, error) {

	// Without one there is nothing to ask with, and the endpoint would reject
	// every attempt — which is how an empty token turns into a request flood.
	if refreshtoken == "" {
		return nil, &OAuthError{Code: "invalid_grant", Desc: "no refresh token stored for this alias"}
	}

	params := url.Values{
		"client_id":     []string{m.ClientID},
		"client_secret": []string{m.ClientSecret},
		"grant_type":    []string{"refresh_token"},
		"refresh_token": []string{refreshtoken},
	}

	endpoint := m.tokenURL
	if endpoint == "" {
		endpoint = tokenEndpoint
	}

	request, err := http.NewRequest("POST", endpoint, bytes.NewReader([]byte(params.Encode())))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// fmt.Println(Dump(request))

	resp, err := m.authclient.httpclient.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	defer io.Copy(io.Discard, resp.Body)

	var target Token
	if resp.StatusCode < 300 {

		err := json.NewDecoder(resp.Body).Decode(&target)
		if err != nil {
			return nil, err
		}

		return &target, nil
	}

	var errResp struct {
		Desc string `json:"error_description"`
		Err  string `json:"error"`
	}
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	if err != nil {
		return nil, fmt.Errorf("token endpoint returned HTTP %d with an unreadable body: %w", resp.StatusCode, err)
	}

	return nil, &OAuthError{Code: errResp.Err, Desc: errResp.Desc}
}

type TokenStore interface {
	LoadTokens() ([]map[string]string, error)
	StoreToken(token *Token) error
}
