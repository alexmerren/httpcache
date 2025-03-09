package httpcache

import (
	"errors"
	"net/http"
)

var (
	// ErrMissingCache will be returned if cache is not set when creating an
	// instance of [Transport].
	ErrMissingCache = errors.New("cache not set when creating transport")

	// ErrMissingConfig will be returned if config is not set when creating an
	// instance of [Transport].
	ErrMissingConfig = errors.New("config not set when creating transport")
)

// Transport is the main interface of the package. It uses [Cache] to persist
// HTTP response, and and uses configuration values from [Config] to interpret
// requests and responses.
type Transport struct {

	// transport handles HTTP requests. This is hardcoded to be
	// [http.DefaultTransport].
	transport http.RoundTripper

	// cache handles persisting HTTP responses.
	cache Cache

	// config describes which HTTP responses to cache and how they are cached.
	config *Config
}

// NewTransport creates a [Transport]. If the cache is nil, return
// [ErrMissingCache]. If the config is nil, return [ErrMissingConfig].
func NewTransport(config *Config, cache Cache) (*Transport, error) {
	if cache == nil {
		return nil, ErrMissingCache
	}
	if config == nil {
		return nil, ErrMissingConfig
	}

	return &Transport{
		transport: http.DefaultTransport,
		cache:     cache,
		config:    config,
	}, nil
}

// RoundTrip wraps the [http.DefaultTransport] RoundTrip to execute HTTP
// requests and persists, if necessary, the responses. If [Cache] returns
// [ErrNoResponse], then execute a HTTP request and persist the response if
// passing the criteria in [Transport.shouldSaveResponse].
func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.cache.Read(request)
	if err == nil {
		return response, nil
	}

	if !errors.Is(err, ErrNoResponse) {
		return nil, err
	}

	response, err = t.transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if t.shouldSaveResponse(response.StatusCode, response.Request.Method) {
		err = t.cache.Save(response)
		if err != nil {
			response.Body.Close()
			return nil, err
		}
	}

	return response, nil
}

// shouldSaveResponse is responsible for interpreting configuration values from
// [Config] to determine if a HTTP response should be persisted. Any new values
// added to [Config] can be used here as criteria.
func (t *Transport) shouldSaveResponse(statusCode int, method string) bool {
	isAllowedStatusCode := contains(t.config.AllowedStatusCodes, statusCode)
	isAllowedMethod := contains(t.config.AllowedMethods, method)

	return isAllowedStatusCode && isAllowedMethod
}

func contains[T comparable](slice []T, searchValue T) bool {
	for _, value := range slice {
		if value == searchValue {
			return true
		}
	}
	return false
}
