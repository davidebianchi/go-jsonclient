package jsonclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Headers map. A key/value map of headers
type Headers map[string]string

// Client basic structure
type Client struct {
	BaseURL        *url.URL
	DefaultHeaders Headers

	client *http.Client
}

// Options to pass to create a new client
type Options struct {
	BaseURL string
	Headers Headers
}

// New function create a client using passed options
// BaseURL must have a trailing slash
func New(opts Options) (*Client, error) {
	fmt.Printf("ASDASDASD %+v", opts)
	baseURL, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		BaseURL:        baseURL,
		DefaultHeaders: Headers{},

		client: http.DefaultClient,
	}

	if opts.Headers != nil {
		client.DefaultHeaders = opts.Headers
	}

	return client, nil
}

// NewRequestWithContext function works like the function of `net/http` package,
// simplified to a easier use with json request.
// Context and method are handled in the same way of core `net/http` package.
// Main differences are:
// * `urlString` params accept a string, that is converted to a type `url.URL`
// adding the client BaseURL (if provided in client options). To works correctly,
// urlString should not starts with `/` (otherwise BaseURL is not set)
// * body params in converted to a `json` buffer. If body is passed, the header
// `Content-Type: application/json` is automatically added to the request.
//
// To the request are added all the DefaultHeaders (if body is passed,
// `application/json` content-type takes precedence over DefaultHeaders).
func (c *Client) NewRequestWithContext(ctx context.Context, method string, urlStr string, body interface{}) (*http.Request, error) {
	if c.BaseURL.Path != "" && !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buffer io.ReadWriter
	if body != nil {
		buffer = &bytes.Buffer{}
		enc := json.NewEncoder(buffer)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buffer)
	if err != nil {
		return nil, err
	}

	for k, v := range c.DefaultHeaders {
		req.Header.Set(k, v)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// NewRequest function is same of NewRequestWithContext, without context
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	return c.NewRequestWithContext(context.Background(), method, urlStr, body)
}

// Do function executes http request using the passed request.
// This function automatically handles response in json to be decoded and saved
// into the `v` param.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		default:
		}
		return nil, err
	}
	defer resp.Body.Close()

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err := json.NewDecoder(resp.Body).Decode(v)
			if err != nil && err != io.EOF {
				return nil, err
			}
		}
	}

	return resp, nil
}
