# httpcache

[![Go Report Card](https://goreportcard.com/badge/github.com/alexmerren/httpcache)](https://goreportcard.com/report/github.com/alexmerren/httpcache)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.21-61CFDD.svg?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/alexmerren/httpcache.svg)](https://pkg.go.dev/github.com/alexmerren/httpcache)

httpcache is a local cache for HTTP requests and responses, wrapping `http.RoundTripper` from Go standard library.

## Features

httpcache has a few useful features:

- Store and retrieve HTTP responses for any type of request;
- Expire responses after a customisable time duration;
- Decide when to store responses based on status code and request method.

If you want to request a feature then please open a [GitHub Issue](https://www.github.com/alexmerren/httpcache/issues) today!

## Quick Start

This module can be installed using the command line:

```bash
go get -u github.com/alexmerren/httpcache
```

Here's an example of using the `httpcache` module to cache responses:

```go
func main() {
	// Create a new SQLite database to store HTTP responses.
	cache, _ := httpcache.NewSqliteCache("database.sqlite")

	// Create a config with a behaviour of:
	// 	- Storing responses with status code 200;
	// 	- Storing responses from HTTP requests using method "GET";
	// 	- Expiring responses after 7 days...
	config := httpcache.NewConfigBuilder().
		WithAllowedStatusCodes([]int{http.StatusOK}).
		WithAllowedMethods([]string{http.MethodGet}).
		WithExpiryTime(time.Duration(60*24*7) * time.Minute).
		Build()

	// ... or use the default config.
	config = httpcache.DefaultConfig

	// Create a transport with the SQLite cache and config.
	cachedTransport, _ := httpcache.NewTransport(config, cache)

	// Create a HTTP client with the cached roundtripper.
	httpClient := http.Client{
		Transport: cachedTransport,
	}

	// Do first request to populate local database.
	httpClient.Get("https://www.google.com")

	// Subsequent requests read from database with no outgoing HTTP request.
	for _ = range 10 {
		response, _ := httpClient.Get("https://www.google.com")
		defer response.Body.Close()
		responseBody, _ := io.ReadAll(response.Body)

		fmt.Println(string(responseBody))
	}
}
```

## ❓ Questions and Support

Any questions can be submitted via [GitHub Issues](https://www.github.com/alexmerren/httpcache/issues). Feel free to start contributing or asking any questions required!
