package jsonclient

import (
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

	raw string
}

// ErrHTTP define an http error
var ErrHTTP = errors.New("http error")

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%v %v: %d - %+v",
		e.Response.Request.Method,
		e.Response.Request.URL,
		e.Response.StatusCode,
		e.raw,
	)
}
func (e *HTTPError) Unwrap() error {
	return e.Err
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
		errorData.raw = string(data)
	}
	return errorData
}
