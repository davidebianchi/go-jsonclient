<div align="center">

# Go Json Client

[![Build Status][github-actions-svg]][github-actions]
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]
[![Coverage Status][coveralls-svg]][coveralls-link]

</div>

Go Json Client simplify the http request in json.

It uses `net/http` core package as http client.

## Install

This library require golang at version >= 1.13

```sh
go get -u github.com/davidebianchi/go-jsonclient
```

## Example usage

### Make a request

If you want to create a json client to call a specific BaseUrl with default
authentication headers:

```go
func handleRequest () {
  opts := jsonclient.Options{
    BaseURL: "http://base-url:8080/api/url/",
    Headers: jsonclient.Headers{
      "some":  "header",
      "other": "value",
    },
  }
  client, err := jsonclient.New(opts)
  if err != nil {
    panic("Error creating client")
  }

  var data = map[string]interface{}{
    "some": "json format",
    "foo":  "bar",
    "that": float64(3),
  }
  req, err := client.NewRequest(http.MethodPost, "my/path", data)
  if err != nil {
    panic("Error creating request")
  }

  type Response struct {
    my string
  }
  v := Response{}
  // server response is: {"my": "data"}
  response, err := client.Do(req, &v)
  if err != nil {
    panic("Error making request")
  }

  if Response.my != "data" {
    panic("response data is not mine")
  }
}
```

The library also check the status code of the request. If status code si not 2xx, it will return an `HTTPError`.

## API

### Accepted client options

In the `New` function, it is possible to add some options. None of the following options are required.

* **BaseURL**: set the base url. BaseURL must be absolute and starts with `http` or `https` scheme. It must end with a trailing slash `/`. Example of valid BaseUrl: `"http://base-url:8080/api/url/"`
* **Headers**: a map of headers to add to all the requests. For example, it could be useful when it is required an auth header.
* **HTTPClient** (default to `http.DefaultClient`): an http client to use instead of the default http client. It could be useful for example for testing purpose.

## Versioning

We use [SemVer][semver] for versioning. For the versions available,
see the [tags on this repository](https://github.com/davidebianchi/go-jsonclient/tags).

[github-actions]: https://github.com/davidebianchi/go-jsonclient/actions
[github-actions-svg]: https://github.com/davidebianchi/go-jsonclient/workflows/Test%20and%20build/badge.svg
[godoc-svg]: https://godoc.org/github.com/davidebianchi/go-jsonclient?status.svg
[godoc-link]: https://pkg.go.dev/github.com/davidebianchi/go-jsonclient
[go-report-card]: https://goreportcard.com/badge/github.com/davidebianchi/go-jsonclient
[go-report-card-link]: https://goreportcard.com/report/github.com/davidebianchi/go-jsonclient
[semver]: https://semver.org/
[coveralls-svg]: https://coveralls.io/repos/github/davidebianchi/go-jsonclient/badge.svg?branch=master
[coveralls-link]: https://coveralls.io/github/davidebianchi/go-jsonclient?branch=master
