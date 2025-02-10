package httpcache

import (
	"errors"
	"net/http"
)

// Add doc
var ErrNoCache = errors.New("no cache set")

// Add doc
type Transport struct {

	// Add doc
	transport http.RoundTripper

	// Add doc
	cache Cache

	// Add doc
	config *Config
}

// Add doc
func NewTransport(config *Config) (*Transport, error) {
	transport := &Transport{
		transport: http.DefaultTransport,
		cache:     config.Cache,
		config:    config,
	}

	if config.Cache == nil {
		return nil, ErrNoCache
	}

	return transport, nil
}

// Add doc
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

	if !t.shouldSaveResponse(response.StatusCode, response.Request.Method) {
		return response, nil
	}

	err = t.cache.Save(response)
	if err != nil {
		response.Body.Close()
		return nil, err
	}

	return response, nil
}

func (t *Transport) shouldSaveResponse(statusCode int, method string) bool {
	isAllowedStatusCode := contains(*t.config.AllowedStatusCodes, statusCode)
	isAllowedMethod := contains(*t.config.AllowedMethods, method)

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
