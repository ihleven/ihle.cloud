package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ihleven/ihle.cloud/pkg/hi"
)

func New(issuer, secretkey string, duration int, cookieName string, cookieSameSite http.SameSite, store hi.TokenStore, authstore authstore) {

	ctx, cncl := context.WithCancel(context.Background())

	AuthenticatorPKG = Authenticator{
		ctx:  ctx,
		cncl: cncl,
		// m:        map[string]*TokenRefresher{},
		accounts:  map[string]*Account{},
		tokenmngr: hi.NewTokenMngr(store),
		JWT: JWT{
			Issuer:    issuer,
			SecretKey: secretkey,
			Duration:  time.Duration(duration) * time.Second,
		},
		CookieConfig: CookieConfig{
			Name:     cookieName,
			SameSite: cookieSameSite,
		},
		authstore: authstore,
	}

	accounts, err := authstore.LoadAccounts()
	if err != nil {
		log.Fatal(err)
	}

	for _, a := range accounts {
		AuthenticatorPKG.accounts[a.ID] = &a
	}

	// for k, v := range LoadTokens() {
	// 	AuthenticatorPKG.m[k] = newTokenRefresher(context.Background(), *v, time.Now())
	// }

	fmt.Println("init:", AuthenticatorPKG.accounts)
	// fmt.Println("init:", AuthenticatorPKG.m)
}

// AuthenticatorPKG ist die zentrale Auth-Instanz auf Package-Level, die die Accounts (backend/accounts) und Tokens (backend/hitokens) verwaltet.
var AuthenticatorPKG Authenticator

type Authenticator struct {
	sync.RWMutex
	// config     Config
	ctx  context.Context
	cncl context.CancelFunc
	// httpclient *http.Client
	// m          map[string]*TokenRefresher
	accounts map[string]*Account

	authstore

	tokenmngr *hi.TokenMngr

	JWT
	CookieConfig
}
type CookieConfig struct {
	Name     string
	SameSite http.SameSite
}

func (a *Authenticator) Account(id string) *Account {
	account, ok := a.accounts[id]
	if !ok {
		return a.accounts["anonymous"]
	}
	return account
}

func (a *Authenticator) HasValidHiToken(account *Account) bool {

	if refr := a.tokenmngr.GetTokenRefresher(account.Settings.Hidrive.Alias); refr == nil {
		return false
	}
	return true
}

func (a *Authenticator) GetToken(account *Account) (string, error) {

	if refr := a.tokenmngr.GetTokenRefresher(account.Settings.Hidrive.Alias); refr != nil {
		return refr.GetAccessToken()
	}
	return "", errors.New("token for " + account.ID + " not found")
}

func GetAccountUnused(r *http.Request) (*Account, error) {

	cookie, err := r.Cookie("jwt")
	if err != nil {
		return nil, err
	}

	claims, err := AuthenticatorPKG.ParseClaims(cookie.Value)
	if err != nil {
		return nil, err
	}

	AuthenticatorPKG.RLock()
	defer AuthenticatorPKG.RUnlock()

	account, ok := AuthenticatorPKG.accounts[claims.Subject]
	if !ok {
		return nil, errors.New("account %s not found: " + account.ID)
	}
	fmt.Printf("GetAccount: %s\n", account.ID)
	return account, nil
}

// func (a *Authenticator) add(authdata hiauth.Token, account string) error {
// 	a.Lock()
// 	defer a.Unlock()

// 	a.m[account] = newTokenRefresher(a.ctx, authdata, time.Time{})
// 	return nil
// }

// func (a *Authenticator) GetAccessTokenDep(id string) (string, error) {
// 	a.RLock()
// 	defer a.RUnlock()

// 	account, ok := a.accounts[id]
// 	if !ok {
// 		return "", fmt.Errorf("account not found: %s", id)
// 	}

// 	tokenrefresher, ok := a.m[account.Settings.Hidrive.Alias]
// 	if !ok {
// 		return "", fmt.Errorf("authdata not found: %s", account.Settings.Hidrive.Alias)
// 	}

// 	accesstoken, err := tokenrefresher.GetAccessToken()
// 	return accesstoken, err

// }

// func newTokenRefresher(ctx context.Context, token hiauth.Token, expiration time.Time) *TokenRefresher {
// 	a := &TokenRefresher{
// 		stream: make(chan tokenResponse),
// 		token:  token,
// 	}
// 	go a.refreshloop(ctx, token, expiration)
// 	return a
// }

// type TokenRefresher struct {
// 	stream chan tokenResponse
// 	token  hiauth.Token
// }

// type tokenResponse struct {
// 	token hiauth.Token
// 	Err   error
// }

// func (a *TokenRefresher) GetAccessToken() (string, error) {
// 	if a == nil {
// 		return "", errors.New("nil refresher")
// 	}
// 	tokenresponse := <-a.stream
// 	// 	return t.token.AccessToken, t.Err
// 	// workaround
// 	return tokenresponse.token.AccessToken, tokenresponse.Err
// }

// func (a *TokenRefresher) TokenDep() hiauth.Token {
// 	if a == nil {
// 		return hiauth.Token{}
// 	}
// 	tokenresponse := <-a.stream
// 	// 	return t.token.AccessToken, t.Err
// 	// workaround
// 	return tokenresponse.token
// }

// const (
// 	lifeSpanSafetyMargin = 59 * time.Minute // 10 * time.Millisecond
// 	retryDelay           = 11 * time.Millisecond
// )

// func (a *TokenRefresher) refreshloop(ctx context.Context, hitoken hiauth.Token, expiresAt time.Time) {

// 	var expiration time.Duration
// 	var err error

// 	// expiration := time.Duration(a.token.ExpiresIn) * time.Second
// 	if !expiresAt.IsZero() {
// 		expiration = time.Until(expiresAt)
// 		fmt.Println("expiration:", expiration)
// 	}
// 	expired := time.After(expiration - lifeSpanSafetyMargin)
// 	// time.NewTimer

// 	fmt.Println("expiration:", expiration)

// 	client := hiauth.NewClient(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))

// 	for {
// 		select {
// 		case a.stream <- tokenResponse{token: a.token, Err: err}:

// 		case <-expired:

// 			log.Println("Token expired")
// 			t, e := client.RefreshToken(a.token.RefreshToken)
// 			if e != nil {
// 				log.Println("Error refreshing token:", err)
// 				expiration = retryDelay
// 				err = e
// 			} else {
// 				expiration = time.Duration(t.ExpiresIn) * time.Second
// 				a.token = *t
// 				StoreToken(t)
// 			}
// 			expired = time.After(expiration - lifeSpanSafetyMargin)
// 			fmt.Println("expiration", expiration, "expired", expired, "token:", t.ExpiresIn)
// 		case <-ctx.Done():
// 			return
// 		}
// 	}
// }

// func GetPrefixAndToken(r *http.Request) (string, string, error) {

// 	start := time.Now()
// 	cookie, err := r.Cookie("jwt")
// 	if err != nil {
// 		fmt.Println("cookie error:", err.Error())
// 		return "", "", err
// 	}

// 	claims, err := AuthenticatorPKG.VerifyJWT(cookie.Value)
// 	if err != nil {
// 		fmt.Println(err.Error())

// 		return "", "", err
// 	}

// 	// accesstoken, err := AuthenticatorPKG.GetAccessToken(claims.Subject)
// 	// if err != nil {
// 	// 	fmt.Println(err.Error())

// 	// 	return "", ""
// 	// }

// 	//
// 	// return "", accesstoken

// 	AuthenticatorPKG.RLock()
// 	defer AuthenticatorPKG.RUnlock()

// 	account, ok := AuthenticatorPKG.accounts[claims.Subject]
// 	if !ok {
// 		return "", "", errors.New("account %s not found: " + account.ID)
// 	}

// 	tokenrefresher, ok := AuthenticatorPKG.m[account.Settings.Hidrive.Alias]
// 	if !ok {
// 		return "", "", errors.New("token %s not found: " + account.ID)
// 	}

// 	accesstoken, err := tokenrefresher.GetAccessToken()
// 	fmt.Println("auth duration:", time.Since(start))

// 	return account.Settings.Hidrive.Home, accesstoken, nil
// }
