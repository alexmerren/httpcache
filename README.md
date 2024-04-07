# HTTPcache

[![Go Report Card](https://goreportcard.com/badge/github.com/alexmerren/httpcache)](https://goreportcard.com/report/github.com/alexmerren/httpcache)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.21-61CFDD.svg?style=flat-square)
![Build Passing](https://github.com/alexmerren/httpcache/actions/workflows/go.yml/badge.svg)

HTTPcache is a fast, local cache for HTTP requests and responses and wraps the default `http.RoundTripper` from Go standard library.

## Features

HTTPcache has a few useful features:

* Store and retrieve HTTP responses for any type of request.
* Expire responses after a customisable time duration.
* Decide when to store responses based on status code.

If you want to request a feature then open a [GitHub Issue](https://www.github.com/alexmerren/httpcache/issues) today!

## Quick Start

This module can be installed using the command line:

```bash
go get -u github.com/alexmerren/httpcache
```

Here's an example of using the `httpcache` module to cache responses:

```go
package main

func main() {
    // Create a new cached round tripper that:
    // * Only stores responses with status code 200.
    // * Refuses to store responses with status code 404. 
    cache := httpcache.NewCachedRoundTripper(
        httpcache.WithAllowedStatusCodes([]int{200}),
        httpcache.WithDeniedStatusCodes([]int{404}),
    )

    // Create HTTP client with cached round tripper.
    httpClient := &http.Client{
        Transport: cache,
    }

    // Do first request to populate local database.
    httpClient.Get("https://www.google.com") 

    // Subsequent requests read from database with no outgoing HTTP request.
    for _ = range 10 {
        response, _ = httpClient.Get("https://www.google.com")
        defer response.Body.Close()
        responseBody = io.ReadAll(response.body)

        fmt.Println(responseBody)
    }
}
```

## ❓ Questions and Support

Any questions can be submitted via [GitHub Issues](https://www.github.com/alexmerren/httpcache/issues). Feel free to start contributing or asking any questions required!
