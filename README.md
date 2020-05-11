<div align="center">

# Go Json Client

[![Build Status][github-actions-svg]][github-actions]
[![Go Report Card][go-report-card]][go-report-card-link]
[![GoDoc][godoc-svg]][godoc-link]

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
    BasePath: apiURL,
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
