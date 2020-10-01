package jsonclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const apiURL = "https://base-url:8080/api/url/"

var baseURL = &url.URL{
	Path:   "/api/url/",
	Host:   "base-url:8080",
	Scheme: "https",
}

func TestNewClient(t *testing.T) {
	t.Run("correctly returns client", func(t *testing.T) {
		opts := Options{BaseURL: apiURL}
		client, err := New(opts)

		require.NoError(t, err, "create client error")
		require.Exactly(t, client, &Client{
			BaseURL:        baseURL,
			DefaultHeaders: Headers{},

			client: http.DefaultClient,
		})
	})

	t.Run("correctly returns client with empty options", func(t *testing.T) {
		opts := Options{}
		client, err := New(opts)
		require.NoError(t, err, "create client error")
		require.Exactly(t, client, &Client{
			BaseURL:        &url.URL{},
			DefaultHeaders: Headers{},

			client: http.DefaultClient,
		})
	})

	t.Run("correctly returns client with custom headers and http client", func(t *testing.T) {
		headers := map[string]string{"h1": "v1", "h2": "v2"}
		customHTTPClient := &http.Client{
			Timeout: 1234,
		}

		opts := Options{
			Headers: headers,
			Client:  customHTTPClient,
		}
		client, err := New(opts)
		require.NoError(t, err, "create client error")
		require.Exactly(t, client, &Client{
			BaseURL:        &url.URL{},
			DefaultHeaders: headers,

			client: customHTTPClient,
		})
	})

	t.Run("throws if base url is not correct", func(t *testing.T) {
		opts := Options{
			BaseURL: "/not\tcorrect",
		}
		client, err := New(opts)

		require.True(t, strings.Contains(err.Error(), "invalid control character in URL"))
		require.Nil(t, client, "client is not nil")
	})

	t.Run("throws if base url is not absolute", func(t *testing.T) {
		opts := Options{
			BaseURL: "/notcorrect",
		}
		client, err := New(opts)

		require.EqualError(t, err, "baseURL should be an absolute url", "create client error")
		require.Nil(t, client, "client is not nil")
	})

	t.Run("throws if base url has unsupported scheme", func(t *testing.T) {
		opts := Options{
			BaseURL: "ws://notcorrect",
		}
		client, err := New(opts)

		require.EqualError(t, err, "unsupported scheme: ws", "create client error")
		require.Nil(t, client, "client is not nil")
	})

	t.Run("throws if base url doesn't end with /", func(t *testing.T) {
		opts := Options{
			BaseURL: strings.TrimSuffix(apiURL, "/"),
		}
		client, err := New(opts)

		require.EqualError(t, err, "BaseURL must end with a trailing slash")
		require.Nil(t, client, "client is not nil")
	})
}

func TestNewRequestWithContext(t *testing.T) {
	opts := Options{
		BaseURL: apiURL,
		Headers: Headers{
			"some":  "header",
			"other": "value",
		},
	}
	client, err := New(opts)
	require.NoError(t, err, "create client error")

	type testKeyCtx struct{}
	contextValue := "context-value"
	ctx := context.WithValue(context.Background(), testKeyCtx{}, contextValue)

	t.Run("throws if url parsing throw", func(t *testing.T) {
		req, err := client.NewRequestWithContext(context.Background(), http.MethodGet, "\t", nil)

		require.True(t, strings.Contains(err.Error(), "invalid control character in URL"))
		require.Nil(t, req, "req is not nil")
	})

	t.Run("throws if baseURL and urlStr are absolute", func(t *testing.T) {
		req, err := client.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.org", nil)

		require.EqualError(t, err, "baseURL and urlStr cannot be both absolute")
		require.Nil(t, req, "req is not nil")
	})

	t.Run("correctly create request path", func(t *testing.T) {
		req, err := client.NewRequestWithContext(ctx, http.MethodGet, "my-resource", nil)

		require.NoError(t, err, "new request not errors")
		require.Exactly(t, "https://base-url:8080/api/url/my-resource", req.URL.String())
		require.Exactly(t, req.Header.Get("Content-Type"), "")
		v := req.Context().Value(testKeyCtx{})
		require.Exactly(t, contextValue, v, "context is not correct")
	})

	t.Run("correctly create request path with query params", func(t *testing.T) {
		req, err := client.NewRequestWithContext(ctx, http.MethodGet, "my-resource?query=params", nil)

		require.NoError(t, err, "new request not errors")
		require.Exactly(t, "https://base-url:8080/api/url/my-resource?query=params", req.URL.String())
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
		client, err := New(opts)
		require.NoError(t, err, "create client error")

		req, err := client.NewRequestWithContext(ctx, http.MethodPost, "https://local-server/my-resource", data)
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
			BaseURL: apiURL,
			Headers: Headers{
				"Content-Type": "not-a-json",
			},
		}
		client, err := New(opts)
		require.NoError(t, err, "create client error")

		req, err := client.NewRequestWithContext(ctx, http.MethodPost, "my-resource", data)
		require.NoError(t, err, "request error")

		require.Exactly(t, req.Header.Get("Content-Type"), "application/json")
	})
}

func TestNewRequest(t *testing.T) {
	opts := Options{BaseURL: apiURL}
	client, err := New(opts)
	require.NoError(t, err, "create client error")

	t.Run("correctly create request path", func(t *testing.T) {
		req, err := client.NewRequest(http.MethodGet, "my-resource", nil)

		require.NoError(t, err, "new request not errors")
		require.Exactly(t, "https://base-url:8080/api/url/my-resource", req.URL.String())
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
	opts := Options{BaseURL: apiURL}
	client, err := New(opts)
	require.NoError(t, err, "create client error")

	type Response struct {
		Message string `json:"message"`
	}
	response := `{"message": "my response"}`

	setupServer := func(responseBody string, statusCode int) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(statusCode)
			if responseBody != "" {
				w.Write([]byte(responseBody))
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
		statusCode := 404
		responseBody := `{"message":"Not Found","error":"Not Found","statusCode":404}`
		s := setupServer(responseBody, statusCode)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		v := Response{}
		resp, err := client.Do(req, &v)
		require.Nil(t, resp)
		require.Exactly(t, Response{}, v, "v not empty")
		require.Error(t, err, "response error")

		expectedError := fmt.Sprintf("GET %s/: 404 - %s", s.URL, responseBody)
		require.EqualError(t, err, expectedError, "error not correct")
		var e *HTTPError
		require.True(t, errors.As(err, &e))
		require.Equal(t, 404, e.StatusCode, "error returning status code")
	})

	t.Run("correctly return response if http error expecting array", func(t *testing.T) {
		s := setupServer(`{"message": "Bad Request"}`, 404)
		defer s.Close()

		requestURL := fmt.Sprintf("%s/", s.URL)
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		require.NoError(t, err, "wrong request creation")

		v := []Response{}
		resp, err := client.Do(req, &v)
		require.Nil(t, resp)
		require.Exactly(t, []Response{}, v, "v not empty")

		expectedError := fmt.Sprintf(`GET %s/: 404 - {"message": "Bad Request"}`, s.URL)
		require.EqualError(t, err, expectedError, "response error")
		var e *HTTPError
		require.True(t, errors.As(err, &e))
		require.Equal(t, 404, e.StatusCode, "error returning status code")
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

func TestIntegration(t *testing.T) {
	setupServer := func(responseBody, expectedRequestBody string, statusCode int) *httptest.Server {
		t.Helper()
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()

			if expectedRequestBody != "" {
				bodyBytes, _ := ioutil.ReadAll(req.Body)
				require.Equal(t, expectedRequestBody, strings.TrimSuffix(string(bodyBytes), "\n"))
			}

			w.WriteHeader(statusCode)
			if responseBody != "" {
				w.Write([]byte(responseBody))
				return
			}
			w.Write(nil)
		}))
	}

	t.Run("get", func(t *testing.T) {
		expectedMessage := "my message"
		type Response struct {
			Message string `json:"message"`
		}
		s := setupServer(fmt.Sprintf(`{"message": "%s"}`, expectedMessage), "", 200)
		opts := Options{
			BaseURL: fmt.Sprintf("%s/api/", s.URL),
		}
		client, err := New(opts)
		require.NoError(t, err, "throws create client")

		req, err := client.NewRequest(http.MethodGet, "/my-resource", nil)
		require.NoError(t, err, "throws creating request")

		response := Response{}
		r, err := client.Do(req, &response)
		require.NoError(t, err, "throws exec request")

		require.Equal(t, Response{
			Message: expectedMessage,
		}, response)
		require.Equal(t, 200, r.StatusCode, "wrong status code")
	})

	t.Run("post", func(t *testing.T) {
		myID := "my id"
		type Response struct {
			ID string `json:"id"`
		}
		type RequestBody struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}

		expectedName := "my name"
		expectedDescription := "my description"
		expectedRequestBody := fmt.Sprintf(`{"name":"%s","description":"%s"}`, expectedName, expectedDescription)
		s := setupServer(fmt.Sprintf(`{"id": "%s"}`, myID), expectedRequestBody, 200)
		opts := Options{
			BaseURL: fmt.Sprintf("%s/api/", s.URL),
		}
		client, err := New(opts)
		require.NoError(t, err, "throws create client")

		requestBody := RequestBody{
			Name:        expectedName,
			Description: expectedDescription,
		}
		req, err := client.NewRequest(http.MethodPost, "/my-resource", requestBody)
		require.NoError(t, err, "throws creating request")
		response := Response{}
		r, err := client.Do(req, &response)
		require.NoError(t, err, "throws exec request")

		require.Equal(t, Response{
			ID: myID,
		}, response)
		require.Equal(t, 200, r.StatusCode, "wrong status code")
	})
}
