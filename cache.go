package httpcache

import (
	"bytes"
	"errors"
	"io"
	"net/http"
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
	defaultCacheStore = NewDefaultResponseStore()
)

type CachedRoundTripper struct {
	roundTripper       http.RoundTripper
	cacheStore         ResponseStorer
	deniedStatusCodes  []int
	allowedStatusCodes []int
}

func NewCachedRoundTripper(options ...func(*CachedRoundTripper)) *CachedRoundTripper {
	roundTripper := &CachedRoundTripper{
		roundTripper:       http.DefaultTransport,
		cacheStore:         defaultCacheStore,
		deniedStatusCodes:  defaultDeniedStatusCodes,
		allowedStatusCodes: defaultAllowedStatusCodes,
	}

	for _, optionFunc := range options {
		optionFunc(roundTripper)
	}

	return roundTripper
}

func (h *CachedRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := h.cacheStore.Read(request)
	if err == nil {
		return response, nil
	}

	if err != nil && !errors.Is(err, ErrNoResponse) {
		return nil, err
	}

	// Store a copy of the request body so we can retrieve it after calling
	// roundTripper.RoundTrip(request).
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

	// Do() reads the request body, so we reset the request body so that the
	// cache store can read it as part of the composite key.
	response, err = h.roundTripper.RoundTrip(request)
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

	err = h.cacheStore.Create(response)
	if err != nil {
		response.Body.Close()
		response.Request.Body.Close()
		return nil, err
	}

	return response, nil
}
