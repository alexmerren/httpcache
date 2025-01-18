package httpcache

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	defaultDeniedStatusCodes = []int{
		http.StatusNotFound,
		http.StatusBadRequest,
		http.StatusForbidden,
		http.StatusUnauthorized,
		http.StatusMethodNotAllowed,
	}

	defaultAllowedStatusCodes = []int{
		http.StatusOK,
	}

	defaultExpiryTime = time.Duration(0)

	defaultCacheName = "httpcache.sqlite"
)

type CachedRoundTripper struct {
	transport          http.RoundTripper
	store              ResponseStorer
	expiryTime         time.Duration
	deniedStatusCodes  []int
	allowedStatusCodes []int
}

func NewCachedRoundTripper(options ...func(*CachedRoundTripper) error) (*CachedRoundTripper, error) {
	roundTripper := &CachedRoundTripper{
		transport:          http.DefaultTransport,
		deniedStatusCodes:  defaultDeniedStatusCodes,
		allowedStatusCodes: defaultAllowedStatusCodes,
		expiryTime:         defaultExpiryTime,
	}

	for _, optionFunc := range options {
		err := optionFunc(roundTripper)
		if err != nil {
			return nil, err
		}
	}

	if roundTripper.store != nil {
		return roundTripper, nil
	}

	store, err := newSqliteResponseStore(defaultCacheName)
	if err != nil {
		return nil, err
	}
	roundTripper.store = store

	return roundTripper, nil
}

func (h *CachedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := h.store.Read(request)
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

	response, err = h.transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	response.Request.Body = io.NopCloser(bytes.NewReader(requestBody))

	if contains(h.deniedStatusCodes, response.StatusCode) {
		return response, nil
	}

	if !contains(h.allowedStatusCodes, response.StatusCode) {
		return response, nil
	}

	err = h.store.Save(response)
	if err != nil {
		response.Body.Close()
		response.Request.Body.Close()
		return nil, err
	}

	return response, nil
}

func contains(slice []int, searchValue int) bool {
	for index := range slice {
		if searchValue == slice[index] {
			return true
		}
	}
	return false
}
