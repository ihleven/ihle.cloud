package hi

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/ihleven/pkg/errors"
)

var GET = "GET"
var POST = "POST"
var PUT = "PUT"

func newRequest(method, uri string, options ...Option) *Request {
	req := &Request{
		Method:  method,
		URL:     uri,
		Headers: map[string]string{},
		Params:  url.Values{},
	}
	for _, apply := range options {
		apply(req)
	}
	return req
}

type Request struct {
	Method  string
	URL     string
	Params  url.Values
	Headers map[string]string
	// Body    []byte
}

type Option func(*Request)

func bearer(value string) Option {
	return func(req *Request) {
		req.Headers["Authorization"] = "Bearer " + value
	}
}

// Type: string
// The path to a filesystem object.
// Example: /users/example/Music
// The shortest possible path is "/", which will always refer to the topmost directory accessible by the authenticated user. For a regular HiDrive user this is the HiDrive "root". If used with a share access_token it will be the shared directory.
// Note: if used in combination with a pid, this value is not allowed to start with "/".
func path(value string) Option {
	return func(req *Request) {

		if value != "" {
			req.Params.Set("path", value)
		}
	}
}

// Type: string
// A comma-separated list of value types that will be included in the response.
// The performance of the call might be influenced by the amount of information requested.
// Therefore, it is recommended to use a "need to know" approach instead of "get all".
// The default is: path,members.name
func fields(value string) Option {
	return func(req *Request) {
		if value != "" {
			fmt.Println("settign fields", value)
			req.Params.Set("fields", value)
		}
	}
}

func Header(header, value string) Option {
	return func(req *Request) {
		req.Headers[header] = value
	}
}

// func XMLBody(body interface{}) Option {
// 	return func(req *Request) {
// 		req.Headers["Content-Type"] = "application/xml"
// 		req.Body, _ = xml.MarshalIndent(body, "", "    ")
// 	}
// }

func Param(key, value string) Option {
	return func(req *Request) {

		if value != "" {
			req.Params.Set(key, value)
		}
		// for _, v := range values {
		// 	if v != "" {
		// 		req.Params.Add(key, v)
		// 	}
		// }
	}
}

// func ParamValues(key string, values ...string) Option {
// 	return func(req *Request) {
// 		for _, value := range values {
// 			if value != "" {
// 				req.Params.Add(key, value)
// 			}
// 		}
// 	}
// }

func Dump(r *http.Request) string {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return fmt.Sprint("Error dumping request:", err.Error())
	}
	return string(dump)
}

// Exec will perform the request via the given http.Client.
// If the returned error is nil, the Response will contain a non-nil body which the user is expected to close.
func (rq *Request) Exec(client *http.Client) (*http.Response, error) {

	// calculating url
	if len(rq.Params) > 0 {
		rq.URL = "https://api.hidrive.strato.com/2.1" + rq.URL + "?" + rq.Params.Encode()
	}
	// var body io.Reader
	// if len(rq.Body) != 0 {
	// 	body = bytes.NewReader(rq.Body)
	// }
	// preparing *http.Request
	request, err := http.NewRequestWithContext(context.Background(), rq.Method, rq.URL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "couldn‘t create %s %s http request", rq.Method, rq.URL)
	}
	for key, value := range rq.Headers {
		request.Header.Set(key, value)
	}
	fmt.Println(Dump(request))

	// submitting *http.Request
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(request)
	if err != nil {
		// Any returned error will be of type *url.Error. The url.Error
		// value's Timeout method will report true if the request timed out.
		switch {
		case os.IsTimeout(err):
			// HTTP 504 Gateway Timeout
			return nil, errors.NewWithCode(504, "timeout exceeded: %s", err.Error())
		}

		// HTTP 502 Bad Gateway
		return nil, errors.WrapWithCode(err, 502, "HTTP client couldn't Do request: %s -> %s", Dump(request), err.Error())
	}

	return resp, nil
}

func (rq *Request) fetchJSON(client *http.Client) (*meta, *Error) {

	response, err := rq.Exec(client)
	if err != nil {

	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}
	var meta meta
	err = json.Unmarshal(bytes, &meta)
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}
	return &meta, nil
}

type response http.Response

func (r *response) ParseXML(target interface{}) error {

	// response := (*http.Response)(r)

	// make sure response.Body is read to end and closed
	// otherwise connection will _not_ be reused.
	defer r.Body.Close()
	defer io.Copy(io.Discard, r.Body)

	if r.StatusCode == 204 {
		// bytes, err := io.ReadAll(r.Body)
		// if len(bytes) > 0 {
		// err = xml.Unmarshal(bytes, target)
		// }
		return nil
	}

	if target != nil {

		err := xml.NewDecoder(r.Body).Decode(target)
		if err != nil {
			return errors.Wrap(err, "Couldn‘t decode xml response body")
		}
	}
	return nil
}

func (r *response) StatusBytes() (int, string, []byte, error) {

	// response := (*http.Response)(r)

	// make sure response.Body is read to end and closed
	// otherwise connection will _not_ be reused.
	defer r.Body.Close()
	defer io.Copy(io.Discard, r.Body) // nach ReadAll unnötig?!?

	bytes, err := io.ReadAll(r.Body)
	return r.StatusCode, r.Header.Get("Content-Type"), bytes, err
}

func (r *response) DrainAndClose() {

	io.Copy(io.Discard, r.Body)
	r.Body.Close()
}
