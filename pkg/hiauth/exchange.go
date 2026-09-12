package hiauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RefreshTokenWithAuthCode exchanges an authorization code for a token, which is
// how an alias gets its first refresh token.
//
// The parameters go in the body rather than the query string: a URL is recorded
// by proxies, gateways and the provider's own access logs, and this request
// carries the client secret and a code that can be redeemed for one.
func RefreshTokenWithAuthCode(id, secret, authcode string) (*Token, error) {

	if authcode == "" {
		return nil, &OAuthError{Code: "invalid_request", Desc: "empty authorization code"}
	}

	params := url.Values{
		"client_id":     []string{id},
		"client_secret": []string{secret},
		"grant_type":    []string{"authorization_code"},
		"code":          []string{authcode},
	}

	request, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("building the token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("asking the token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var e struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &e); err != nil {
			return nil, fmt.Errorf("token endpoint returned HTTP %d and its reason was unreadable", resp.StatusCode)
		}

		return nil, &OAuthError{Code: e.Error, Desc: e.ErrorDescription}
	}

	var t Token
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("could not read the token response: %w", err)
	}

	return &t, nil
}
