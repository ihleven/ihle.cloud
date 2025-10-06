package hi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

var GET = "GET"
var POST = "POST"

// rq führt alle http-Zugriffe auf das HiDrive aus, wird von den Client-Methoden aufgerufen
func rq(method, path string, params url.Values, options ...func(*http.Request)) (*http.Response, error) {

	request, err := http.NewRequest(method, "https://api.hidrive.strato.com/2.1"+path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, Error{HTTPStatus: 500, Message: "Couldn't create request: " + err.Error()}
	}

	for _, apply := range options {
		apply(request)
	}
	Dump(request)
	resp, err := client.Do(request)
	if err != nil {
		if os.IsTimeout(err) {
			// HTTP 504 Gateway Timeout
			return nil, Error{HTTPStatus: 504, Message: "timeout exceeded: " + err.Error()}
		}

		// HTTP 502 Bad Gateway
		return nil, Error{HTTPStatus: 502, Message: "HTTP client couldn't Do request: " + err.Error()}
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		defer io.Copy(io.Discard, resp.Body)

		err := Error{HTTPStatus: resp.StatusCode}
		decerr := json.NewDecoder(resp.Body).Decode(&err)
		if decerr != nil {
			return nil, Error{HTTPStatus: 500, Message: "Couldn't decode response:" + decerr.Error()}
		}

		return nil, err
	}

	return resp, nil
}

func header(header, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(header, value)
	}
}

func headers(header http.Header) func(*http.Request) {
	return func(r *http.Request) {
		for k, v := range header {
			r.Header[k] = v
		}
	}
}

func bearer(token string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func rangeHeader(rangeFrom, rangeTo int) func(*http.Request) {
	return func(req *http.Request) {
		if rangeFrom == 0 && rangeTo == 0 {
			return
		}
		hd := "bytes=" + strconv.Itoa(rangeFrom) + "-"
		if rangeTo != 0 {
			hd += strconv.Itoa(rangeTo)
		}
		req.Header.Set("Range", hd)
	}
}

func Dump(r *http.Request) string {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return fmt.Sprint("Error dumping request:", err.Error())
	}
	return string(dump)
}

// Type: string
// A comma-separated list of value types that will be included in the response.
// The performance of the call might be influenced by the amount of information requested.
// Therefore, it is recommended to use a "need to know" approach instead of "get all".
// The default is: path,members.name
func fields(value string) func(*http.Request) {
	return func(req *http.Request) {
		if value != "" {
			// req.Params.Set("fields", value)
			req.URL.Query().Set("fields", value)
		}
	}
}

// Type: string
// The path to a filesystem object.
// Example: /users/example/Music
// The shortest possible path is "/", which will always refer to the topmost directory accessible by the authenticated user. For a regular HiDrive user this is the HiDrive "root". If used with a share access_token it will be the shared directory.
// Note: if used in combination with a pid, this value is not allowed to start with "/".
func withPath(value string) func(*http.Request) {
	return func(req *http.Request) {

		if value != "" {
			req.URL.Query().Set("path", value)
		}
	}
}
