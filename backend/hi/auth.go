package hi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func RefreshTokenWithAuthCode(id, secret, authcode string) (*Token, error) {

	params := url.Values{
		"client_id":     []string{id},
		"client_secret": []string{secret},
		"grant_type":    []string{"authorization_code"},
		"code":          []string{authcode},
	}

	request, err := http.NewRequest("POST", "https://my.hidrive.com/oauth2/token?"+params.Encode(), nil)
	if err != nil {
		return nil, Error{HTTPStatus: 500, Message: "Couldn't create request: " + err.Error()}
	}

	resp, err := http.DefaultClient.Do(request)
	fmt.Println("resp:", "err:", err, resp.StatusCode)
	if err != nil {
		if os.IsTimeout(err) {
			// HTTP 504 Gateway Timeout
			return nil, Error{HTTPStatus: 504, Message: "timeout exceeded: " + err.Error()}
		}

		// HTTP 502 Bad Gateway
		return nil, Error{HTTPStatus: 502, Message: "HTTP client couldn't Do request: " + err.Error()}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Error{Message: "read error: " + err.Error()}
	}

	if resp.StatusCode != 200 {
		var e struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		err = json.Unmarshal(body, &e)
		if err != nil {
			fmt.Println("err:", err)
		}
		fmt.Println("ERROR:", e)
		return nil, Error{HTTPStatus: resp.StatusCode, Code: e.Error, Message: e.ErrorDescription}
	}

	var t Token
	err = json.Unmarshal(body, &t)
	if err != nil {
		fmt.Println("err:", err)
		return nil, Error{Message: err.Error()}
	}
	fmt.Println("token:", t)
	return &t, nil
}
