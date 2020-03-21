package jsonclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const apiURL = "/api/url/"

func TestGetClient(t *testing.T) {
	t.Run("correctly returns client", func(t *testing.T) {
		opts := Options{BasePath: apiURL}
		client := New(opts)

		require.Exactly(t, client, &Client{
			BaseURL:        &url.URL{Path: apiURL},
			DefaultHeaders: Headers{},

			client: http.DefaultClient,
		})
	})
}

func TestNewRequestWithContext(t *testing.T) {
	opts := Options{
		BasePath: apiURL,
		Headers: Headers{
			"some":  "header",
			"other": "value",
		},
	}
	client := New(opts)

	type testKeyCtx struct{}
	contextValue := "context-value"
	ctx := context.WithValue(context.Background(), testKeyCtx{}, contextValue)

	t.Run("throws if base url has not trailing slash", func(t *testing.T) {
		baseURL := strings.TrimSuffix(apiURL, "/")
		opts := Options{BasePath: baseURL}
		client := New(opts)

		req, err := client.NewRequestWithContext(context.Background(), http.MethodGet, "my-resource", nil)

		require.EqualError(t, err, fmt.Sprintf(`BaseURL must have a trailing slash, but "%s" does not`, baseURL), "new request not errors")
		require.Nil(t, req)
	})

	t.Run("throws if url parsing throw", func(t *testing.T) {
		req, err := client.NewRequestWithContext(context.Background(), http.MethodGet, "	", nil)

		require.Error(t, err, "error creating url")
		require.Nil(t, req, "req is not nil")
	})

	t.Run("correctly create request path", func(t *testing.T) {
		req, err := client.NewRequestWithContext(ctx, http.MethodGet, "my-resource", nil)

		require.NoError(t, err, "new request not errors")
		require.Exactly(t, "/api/url/my-resource", req.URL.String())
		require.Exactly(t, req.Header.Get("Content-Type"), "")
		v := req.Context().Value(testKeyCtx{})
		require.Exactly(t, contextValue, v, "context is not correct")
	})

	t.Run("correctly create request path with query params", func(t *testing.T) {
		req, err := client.NewRequestWithContext(ctx, http.MethodGet, "my-resource?query=params", nil)

		require.NoError(t, err, "new request not errors")
		require.Exactly(t, "/api/url/my-resource?query=params", req.URL.String())
	})

	t.Run("correctly set request body", func(t *testing.T) {
		var data = map[string]interface{}{
			"some": "json format",
			"foo":  "bar",
			"that": float64(3),
		}

		req, err := client.NewRequestWithContext(ctx, http.MethodPost, "my-resource", data)
		require.NoError(t, err, "request error")

		var reqBody map[string]interface{}
		err = json.NewDecoder(req.Body).Decode(&reqBody)
		require.NoError(t, err, "json marshal error")
		require.Exactly(t, data, reqBody, "wrong request body")
		require.Exactly(t, req.Header.Get("Content-Type"), "application/json")
		v := req.Context().Value(testKeyCtx{})
		require.Exactly(t, contextValue, v, "context is not correct")
	})

	t.Run("correctly set request body without base path", func(t *testing.T) {
		var data = map[string]interface{}{
			"some": "json format",
			"foo":  "bar",
			"that": float64(3),
		}
		opts := Options{
			Headers: Headers{
				"some":  "header",
				"other": "value",
			},
		}
		client := New(opts)

		req, err := client.NewRequestWithContext(ctx, http.MethodPost, "my-resource", data)
		require.NoError(t, err, "request error")

		var reqBody map[string]interface{}
		err = json.NewDecoder(req.Body).Decode(&reqBody)
		require.NoError(t, err, "json marshal error")
		require.Exactly(t, data, reqBody, "wrong request body")
		require.Exactly(t, req.Header.Get("Content-Type"), "application/json")
		v := req.Context().Value(testKeyCtx{})
		require.Exactly(t, contextValue, v, "context is not correct")
	})

	t.Run("correctly add default headers to the request", func(t *testing.T) {
		req, err := client.NewRequestWithContext(ctx, http.MethodPost, "my-resource", nil)
		require.NoError(t, err, "request error")

		require.Exactly(t, req.Header.Get("some"), "header")
		require.Exactly(t, req.Header.Get("other"), "value")
	})

	t.Run("content-type header is overwritten to json if body passed", func(t *testing.T) {
		var data = map[string]interface{}{
			"some": "json format",
			"foo":  "bar",
			"that": float64(3),
		}
		opts := Options{
			BasePath: apiURL,
			Headers: Headers{
				"Content-Type": "not-a-json",
			},
		}
		client := New(opts)

		req, err := client.NewRequestWithContext(ctx, http.MethodPost, "my-resource", data)
		require.NoError(t, err, "request error")

		require.Exactly(t, req.Header.Get("Content-Type"), "application/json")
	})
}

func TestNewRequest(t *testing.T) {
	opts := Options{BasePath: apiURL}
	client := New(opts)

	t.Run("correctly create request path", func(t *testing.T) {
		req, err := client.NewRequest(http.MethodGet, "my-resource", nil)

		require.NoError(t, err, "new request not errors")
		require.Exactly(t, "/api/url/my-resource", req.URL.String())
		require.Exactly(t, req.Header.Get("Content-Type"), "")
	})

	t.Run("correctly set request body", func(t *testing.T) {
		var data = map[string]interface{}{
			"some": "json format",
			"foo":  "bar",
			"that": float64(3),
		}

		req, err := client.NewRequest(http.MethodPost, "my-resource", data)
		require.NoError(t, err, "request error")

		var reqBody map[string]interface{}
		err = json.NewDecoder(req.Body).Decode(&reqBody)
		require.NoError(t, err, "json marshal error")
		require.Exactly(t, data, reqBody, "wrong request body")
		require.Exactly(t, req.Header.Get("Content-Type"), "application/json")
	})
}

func TestDo(t *testing.T) {
	opts := Options{BasePath: apiURL}
	client := New(opts)

	type Response struct {
		Message string `json:"message"`
	}
	response := `{"message": "my response"}`

	setupServer := func(body string, statusCode int) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(statusCode)
			if body != "" {
				w.Write([]byte(body))
				return
			}
			w.Write(nil)
		}))
	}

	t.Run("set body and returns response if request has not body", func(t *testing.T) {
		s := setupServer(response, 200)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		v := Response{}
		resp, err := client.Do(req, &v)
		require.NoError(t, err, "wrong request do")
		require.Exactly(t, Response{Message: "my response"}, v, "wrong response body decode")
		require.Exactly(t, 200, resp.StatusCode)
	})

	t.Run("returns response if response has empty body", func(t *testing.T) {
		s := setupServer("", 200)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		v := Response{}
		resp, err := client.Do(req, &v)
		require.NoError(t, err, "wrong request do")
		require.Exactly(t, Response{}, v)
		require.Exactly(t, 200, resp.StatusCode)
	})

	t.Run("returns response if v is nil", func(t *testing.T) {
		s := setupServer("", 200)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		resp, err := client.Do(req, nil)
		require.NoError(t, err, "wrong request do")
		require.Exactly(t, 200, resp.StatusCode)
	})

	t.Run("set body and returns response if v is io.Writer type", func(t *testing.T) {
		s := setupServer(response, 200)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		var buffer = &bytes.Buffer{}

		resp, err := client.Do(req, buffer)
		require.NoError(t, err, "wrong request do")
		require.Exactly(t, 200, resp.StatusCode)
		require.Exactly(t, response, buffer.String(), "buffer copy fails")
	})

	t.Run("throws if body is not a json", func(t *testing.T) {
		notJSONBody := `{not a correct json}`
		s := setupServer(notJSONBody, 200)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		v := Response{}
		resp, err := client.Do(req, &v)
		require.Nil(t, resp, "response is not nil")
		require.Exactly(t, Response{}, v)
		require.EqualError(t, err, "invalid character 'n' looking for beginning of object key string", "wrong do request due to malformed json")
	})

	t.Run("throws if request context done", func(t *testing.T) {
		s := setupServer("", 200)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		ctx, cancelFunc := context.WithTimeout(context.Background(), 0)
		defer cancelFunc()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		resp, err := client.Do(req, nil)
		require.Nil(t, resp, "response is not nil")
		require.EqualError(t, err, "context deadline exceeded", "wrong do request due to exceeded context")
	})

	t.Run("correctly return response if http error", func(t *testing.T) {
		s := setupServer("", 404)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		v := Response{}
		resp, err := client.Do(req, &v)
		require.Exactly(t, Response{}, v, "v not empty")
		require.NoError(t, err, "response error")
		require.Equal(t, 404, resp.StatusCode, "error returning status code")
	})

	t.Run("throws with closed server", func(t *testing.T) {
		s := setupServer("", 200)
		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")
		s.Close()

		resp, err := client.Do(req, nil)
		require.Error(t, err, "response error")
		require.Nil(t, resp, "response is not nil")
	})
}
