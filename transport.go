package httpcache

import (
	"errors"
	"net/http"
	"time"
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

	// config handles logic on saving and reading HTTP responses.
	config *Config
}

// NewTransport creates a [Transport]. If the cache is nil, return
// [ErrMissingCache]. If the config is nil, return [ErrMissingConfig].
func NewTransport(cache Cache, config *Config) (*Transport, error) {
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
// [ErrNoResponse], then execute a HTTP request and persist the response
func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	result, err := t.cache.Read(request)
	if err != nil && !errors.Is(err, ErrNoResult) {
		return nil, err
	}

	isValidResult := isValidResult(t.config, result)
	if isValidResult {
		return result.response, nil
	}

	resultIsInvalid := result != nil && !isValidResult
	if resultIsInvalid {
		err := t.cache.Delete(result.response)
		if err != nil {
			return nil, err
		}
	}

	response, err := t.transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if ok := isEligibleToSave(t.config, response); !ok {
		return response, nil
	}

	err = t.cache.Save(response)
	if err != nil {
		response.Body.Close()
		return nil, err
	}

	return response, nil
}

func isValidResult(config *Config, result *ReadResult) bool {
	return isAllowedMethod(config, result.response.Request.Method) &&
		isAllowedStatusCode(config, result.response.StatusCode) &&
		isNotExpired(config, result.createdAt)
}

func isEligibleToSave(config *Config, response *http.Response) bool {
	return isAllowedMethod(config, response.Request.Method) &&
		isAllowedStatusCode(config, response.StatusCode)
}

func isNotExpired(config *Config, createdAt *time.Time) bool {
	hasExpiryTime := config.ExpiryTime != nil
	if !hasExpiryTime {
		return true
	}
	return time.Now().Before(createdAt.Add(*config.ExpiryTime))
}

func isAllowedStatusCode(config *Config, statusCode int) bool {
	hasAllowedStatusCodes := config.AllowedStatusCodes != nil
	if !hasAllowedStatusCodes {
		return true
	}
	return contains(*config.AllowedStatusCodes, statusCode)
}

func isAllowedMethod(config *Config, method string) bool {
	hasAllowedMethods := config.AllowedMethods != nil
	if !hasAllowedMethods {
		return true
	}
	return contains(*config.AllowedMethods, method)
}

func contains[T comparable](slice []T, searchValue T) bool {
	for _, value := range slice {
		if value == searchValue {
			return true
		}
	}
}
