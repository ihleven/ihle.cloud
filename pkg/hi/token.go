package hi

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
	"strings"
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
	for _, m := range tokens {
		token := Token{
			AccessToken:  "m[access_token]",
			TokenType:    "Bearer",
			ExpiresIn:    0,
			RefreshToken: m["refresh_token"],
			UserID:       "",
			Alias:        m["alias"],
			Scope:        m["scope"],
		}
		mngr.refreshers[token.Alias] = mngr.newTokenRefresher(token, time.Now())
	}
	return &mngr
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

// func (s *TokenMngr) AddOrReplaceToken(token Token) *TokenRefresher {
// 	expiration := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

// 	s.StoreToken(&token)

// 	s.refreshers[token.Alias] = s.newTokenRefresher(token, expiration)
// 	// AuthenticatorPKG.tokenmngr.AddToken(token)
// 	return s.refreshers[token.Alias]
// }

const (
	lifeSpanSafetyMargin = 1 * time.Minute // 10 * time.Millisecond
	retryDelay           = 100 * time.Millisecond
)

func (s *TokenMngr) newTokenRefresher(token Token, expiresAt time.Time) *TokenRefresher {

	a := &TokenRefresher{
		stream: make(chan tokenResponse),
		token:  token,
	}
	// go a.refreshloop(ctx, token, expiration)

	go func() {

		var expiration time.Duration
		var err error

		// expiration := time.Duration(a.token.ExpiresIn) * time.Second
		if !expiresAt.IsZero() {
			expiration = time.Until(expiresAt)
			// fmt.Println("expiration time.Until(expiresAt):", expiresAt)
		}
		// fmt.Println("expiration:", expiration)
		expired := time.After(expiration - lifeSpanSafetyMargin)
		// time.NewTimer

		// client := s.authclient // .NewClient(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))

		for {
			select {
			case a.stream <- tokenResponse{token: a.token, Err: err}:

			case _ = <-expired:

				// log.Printf("Token for %s expired at %s", a.token.Alias, v)
				t, e := s.RefreshToken(a.token.RefreshToken)
				if e != nil {
					// log.Println("Error refreshing token:", err)
					expiration = retryDelay
					err = e
				} else {
					expiration = time.Duration(t.ExpiresIn) * time.Second
					a.token = *t
					// s.StoreToken(t)
					// fmt.Println()
					// log.Printf("Token for %s refreshed", a.token.Alias)
				}
				expired = time.After(expiration - lifeSpanSafetyMargin)
				// fmt.Println("expiration", expiration, "expired", expired, "token:", t.ExpiresIn)
			case <-s.ctx.Done():
				fmt.Println("CTX CLOSED, RETURNING (", a.token.Alias, ")")
				return
			}
		}
	}()

	return a
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

	params := url.Values{
		"client_id":     []string{m.ClientID},
		"client_secret": []string{m.ClientSecret},
		"grant_type":    []string{"refresh_token"},
		"refresh_token": []string{refreshtoken},
	}

	request, err := http.NewRequest("POST", "https://my.hidrive.com/oauth2/token", bytes.NewReader([]byte(params.Encode())))
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
		return nil, err
	}

	return nil, errors.New(errResp.Err + ": " + errResp.Desc)
}

type TokenStore interface {
	LoadTokens() ([]map[string]string, error)
	StoreToken(token *Token) error
}

func LoadTokensDep() map[string]*Token {

	m := map[string]*Token{}
	files, _ := os.ReadDir("./cli/hitokens/")
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		content, err := os.ReadFile("./cli/hitokens/" + file.Name())
		if err != nil {
			fmt.Println("could not read hitoken file", file.Name(), err.Error())
			continue
		}
		var token Token
		err = json.Unmarshal(content, &token)
		if err != nil {
			fmt.Println("could not parse file", file.Name())
			continue
		}
		m[token.Alias] = &token
	}

	return m
}

func StoreTokenDep(token *Token) error {
	bytes, err := json.MarshalIndent(token, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile("./cli/hitokens/"+token.Alias+".json", bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
