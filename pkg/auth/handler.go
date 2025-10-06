package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
	// "errors"
)

type config struct {
	FrontendRedirect, FrontendURL string
}

var Config config = config{
	FrontendRedirect: "http://localhost:3000/login",
	FrontendURL:      "http://localhost:3000",
}

func Login(w http.ResponseWriter, r *http.Request) error {

	r.ParseMultipartForm(5000)
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	// fmt.Println(HashPassword(password))

	account, err := AuthenticatorPKG.GetUser(username)
	if err != nil {
		fmt.Println("account", err)
		return err
	}
	fmt.Println("account", account)

	if !CheckPasswordHash(password, account.Credentials.Password) {
		return errors.New("invalid credentials")
	}

	claims := AuthenticatorPKG.newClaims(account, "famihlie")
	jwt, err := claims.SignedString(AuthenticatorPKG.SecretKey)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     AuthenticatorPKG.CookieConfig.Name,
		Value:    jwt,
		Path:     "/",
		Expires:  claims.ExpiresAt.Time,
		HttpOnly: true,
		Secure:   true,
		SameSite: AuthenticatorPKG.CookieConfig.SameSite,
	})

	if redirect := r.URL.Query().Get("redirect"); redirect != "" {
		uri, _ := url.Parse(r.Referer())
		uri.Path = redirect
		http.Redirect(w, r, uri.String(), http.StatusFound)
		return nil
	}
	// http.Redirect(w, r, r.Referer(), http.StatusFound)
	// return nil
	return RespondJSON(w, claims)
}

var permissions map[string]string = map[string]string{
	"Rudi":  "djvet",
	"DVD":   "mediathek",
	"other": "other",
}

func TokenAuthHandler(w http.ResponseWriter, r *http.Request) error {

	r.ParseForm()
	permission, ok := permissions[r.FormValue("password")]
	if !ok {
		// http.Error(w, "invalid password", 401)
		return errors.NewWithCode(401, "invalid password")
	}

	bearer := r.Header.Get("Authorization")
	bearer = strings.TrimPrefix(bearer, "Bearer ")
	if bearer == "" {
		cookie, err := r.Cookie("jwt")
		if err == nil {
			bearer = cookie.Value
		}
	}

	claims, err := AuthenticatorPKG.ParseClaims(bearer)
	if err != nil {
		c := AuthenticatorPKG.JWT.newClaims(&Account{ID: "anonymous"})
		claims = &c
		// http.Error(w, err.Error(), 401)
		// return
		// claims = &auth.Claims{
		// 	RegisteredClaims: jwt.RegisteredClaims{
		// 		Issuer:    "ihle.cloud",
		// 		Subject:   "", // ???
		// 		Audience:  []string{},
		// 		ExpiresAt: &jwt.NumericDate{time.Now().Add(24 * time.Hour)},
		// 		ID:        "ihleven", // hidrive username
		// 	},
		// 	Permissions: map[string]string{},
		// }
	}
	claims.Audience = append(claims.Audience, permission)
	claims.Permissions[permission] = struct{}{}

	jwt, _ := claims.SignedString(AuthenticatorPKG.JWT.SecretKey)
	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: jwt,
		Path:  "/",
		// MaxAge:   3600,
		Expires:  claims.ExpiresAt.Time,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	return RespondJSON(w, claims)
}

func Logout(w http.ResponseWriter, r *http.Request) error {
	http.SetCookie(w, &http.Cookie{
		Name:   "jwt",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	return nil
}

func Session(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return err
	}

	claims, err := AuthenticatorPKG.ParseClaims(cookie.Value)
	if err != nil {
		return err
	}

	var response = struct {
		*Claims
		Account *Account
	}{Claims: claims}
	// claims.Permissions = map[string]string{"foo": "bar"}
	if r.URL.Query().Has("account") {
		response.Account = AuthenticatorPKG.accounts[claims.Subject]

	}
	// account, accesstoken, err := GetAccessFromClaims(claims)
	// account := AuthenticatorPKG.Account(claims.Subject)
	// fmt.Println("GetAccessFromClaims", account)
	// token, err := AuthenticatorPKG.GetToken(account)

	// if account != nil {
	// 	token := AuthenticatorPKG.m[account.Settings.Hidrive.Alias]
	// 	response["token"], _ = token.GetAccessToken()
	// }

	return RespondJSON(w, response)
}

func Redirect(w http.ResponseWriter, req *http.Request, target string, err error) error {
	http.Redirect(w, req, Config.FrontendRedirect+"?error="+err.Error(), http.StatusFound)
	return nil
}

func Signin(w http.ResponseWriter, req *http.Request) error {

	req.ParseMultipartForm(5000)
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	redirect := req.URL.Query().Get("redirect")
	domain := strings.Split(req.Host, ":")[0]
	fmt.Println("user", username, password, redirect, req.Referer(), Config.FrontendRedirect, "domain", domain, "-")

	account, err := AuthenticatorPKG.GetUser(username)
	if err != nil {
		// if errors.Is(err, ErrUserNotFound) {
		// 	http.Redirect(w, req.Request, Config.FrontendRedirect+"?error="+err.Error(), http.StatusFound)
		// 	// w.WriteHeader(http.StatusUnauthorized)
		// 	return nil
		// }
		// http.Redirect(w, req.Request, Config.FrontendRedirect+"?error=500", http.StatusFound)
		// // w.WriteHeader(http.StatusInternalServerError)
		// return nil
		return Redirect(w, req, redirect, err)
	}

	if !CheckPasswordHash(password, account.Credentials.Password) {
		return Redirect(w, req, redirect, errors.New("invalid credentials"))
	}

	jwt, err := AuthenticatorPKG.newClaims(account, "http://ihle.cloud").SignedString(AuthenticatorPKG.SecretKey)
	if err != nil {
		return Redirect(w, req, redirect, err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    jwt,
		Path:     "/",
		Domain:   domain,
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   strings.HasPrefix(req.Referer(), "https"),
		SameSite: http.SameSiteLaxMode,
	})

	if AuthenticatorPKG.HasValidHiToken(account) {
		if redirect != "" {

			// http.Redirect(w, req.Request, redirect, http.StatusFound)

			fmt.Fprintf(w, `<!DOCTYPE HTML>
<html lang="en-US">
    <head>
        <meta charset="UTF-8">
        <meta http-equiv="refresh" content="0; url=%s">
        <script type="text/javascript">
            window.location.href = "%s"
        </script>
        <title>Page Redirection</title>
    </head>
    <body>
        If you are not redirected automatically, follow this <a href='%s'>%s</a>.
    </body>
</html>`, redirect, redirect, redirect, redirect)
		} else {

			RespondJSON(w, jwt)
		}
	} else {

		params := url.Values{
			"client_id":     []string{os.Getenv("CLIENT_ID")},
			"response_type": []string{"code"},
			"scope":         []string{"admin,rw"},
			"lang":          []string{"de"},
			"state":         []string{"http://localhost:3000/cmsa"}, //account.Settings.Hidrive.Alias},
			"redirect_uri":  []string{"http://localhost:8000/hi/auth/authcode"},
		}

		http.Redirect(w, req, "https://my.hidrive.com/client/authorize?"+params.Encode(), http.StatusFound)
	}
	return nil
}

func RespondJSON(w http.ResponseWriter, value interface{}) error {
	if value == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}
