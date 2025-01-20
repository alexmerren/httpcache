package httpcache

import (
	"bytes"
	"errors"
	"io"
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

	requestBody := []byte{}
	if request.GetBody != nil {
		body, err := request.GetBody()
		if err != nil {
			return nil, err
		}
		requestBody, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		defer body.Close()
	}

	response, err = t.transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	response.Request.Body = io.NopCloser(bytes.NewReader(requestBody))

	if !t.shouldSaveResponse(response.StatusCode) {
		return response, nil
	}

	err = t.cache.Save(response)
	if err != nil {
		response.Body.Close()
		response.Request.Body.Close()
		return nil, err
	}

	return response, nil
}

func (t *Transport) shouldSaveResponse(statusCode int) bool {
	isDeniedStatusCode := contains(*t.config.DeniedStatusCodes, statusCode)
	isAllowedStatusCode := contains(*t.config.AllowedStatusCodes, statusCode)

	return isDeniedStatusCode || !isAllowedStatusCode
}

func contains(slice []int, searchValue int) bool {
	for index := range slice {
		if searchValue == slice[index] {
			return true
		}
	}
	return false
}
