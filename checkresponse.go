package jsonclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// HTTPError struct define http response with status code not 2xx
type HTTPError struct {
	Response   *http.Response
	StatusCode int
	Err        error
	Raw        []byte
}

// ErrHTTP define an http error
var ErrHTTP = errors.New("http error")

func (e *HTTPError) Error() string {
	var delimiter string
	if len(e.Raw) != 0 {
		delimiter = " - "
	}
	return fmt.Sprintf("%v %v: %d%s%+v",
		e.Response.Request.Method,
		e.Response.Request.URL,
		e.Response.StatusCode,
		delimiter,
		string(e.Raw),
	)
}
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// GetRawAsJSON HTTPError content
func (e *HTTPError) GetRawAsJSON(v interface{}) error {
	return json.Unmarshal(e.Raw, v)
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorData := &HTTPError{
		Response:   r,
		StatusCode: r.StatusCode,
		Err:        ErrHTTP,
	}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorData.Raw = data
	}
	return errorData
}
