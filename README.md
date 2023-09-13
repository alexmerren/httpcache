# httpcache

[![Go Report Card](https://goreportcard.com/badge/github.com/alexmerren/httpcache)](https://goreportcard.com/report/github.com/alexmerren/httpcache)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.21-61CFDD.svg?style=flat-square)
![Build Passing](https://github.com/alexmerren/httpcache/actions/workflows/go.yml/badge.svg)

## 🤔 Rationale

`httpcache` is a fast, local cache for various kinds of HTTP responses and
requests. This package is ideal for data collection applications, as there is
no cost to re-do cacheable API requests.

## 💾 Installation

The project can be installed using the following:

```bash
go get github.com/alexmerren/httpcache
```

## 📝 Example

This module is focused around the `CachedClient`. This structure wraps the
functionality of `http.Client` around a local cache using sqlite. You can
easily include a cached client using the following. The default client includes
a set of status codes that it will only store, and a set that it will refuse to
store.

```go
response, err := httpcache.DefaultClient.Get("https://www.google.com")
if err != nil {
    ...
}

defer response.Body.Close()
responseBody, err := io.ReadAll(response.Body)
if err != nil {
    ...
}

fmt.Println(string(responseBody))
```

To set custom denied and accepted status codes, you can invoke a different factory function:

```go
client := httpcache.NewCachedClient(
    []int{404}, // denied status codes
    []int{200}, // allowed status codes
)

response, err := client.Get("https://www.google.com")
if err != nil {
    ...
}

defer response.Body.Close()
responseBody, err := io.ReadAll(response.Body)
if err != nil {
    ...
}

fmt.Println(string(responseBody))
```

## 📚 Reference

The reference documentation is yet to be written. For now, the API is contained in `cache.go` and `store.go`.

## ❓ Questions and Support

Any questions can be submitted via [GitHub Issues](https://www.github.com/alexmerren/httpcache/issues). Feel free to start contributing or asking any questions required!
