package jsonclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// HTTPErrorResponse struct to define http response with status code not 2xx
type HTTPErrorResponse struct {
	Response *http.Response

	raw string
}

func (r *HTTPErrorResponse) Error() string {
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

	errorData := &HTTPErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorData.raw = string(data)
	}
	return errorData
}
