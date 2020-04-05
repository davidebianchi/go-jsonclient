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

func (r *HTTPError) Error() string {
	return fmt.Sprintf("%v %v: %d - %+v",
		r.Response.Request.Method,
		r.Response.Request.URL,
		r.Response.StatusCode,
		r.raw,
	)
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorData := &HTTPError{
		Response:   r,
		StatusCode: r.StatusCode,
		Err:        errors.New("http error"),
	}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorData.raw = string(data)
	}
	return errorData
}
