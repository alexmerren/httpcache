package httpcache

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

var DefaultClient = NewDefaultClient()

var (
	defaultDeniedResponseCodes = []int{
		http.StatusNotFound,
		http.StatusBadRequest,
		http.StatusForbidden,
		http.StatusUnauthorized,
		http.StatusMethodNotAllowed,
	}

	defaultAcceptedResponseCodes = []int{
		http.StatusOK,
	}
)

type CachedClient struct {
	httpClient         *http.Client
	cacheStore         ResponseStorer
	deniedStatusCodes  []int
	allowedStatusCodes []int
}

func NewDefaultClient() *CachedClient {
	return NewCachedClient(NewDefaultResponseStore(), defaultDeniedResponseCodes, defaultAcceptedResponseCodes)
}

func NewCachedClient(responseStore ResponseStorer, deniedStatusCodes, allowedStatusCodes []int) *CachedClient {
	return &CachedClient{
		httpClient:         http.DefaultClient,
		cacheStore:         responseStore,
		deniedStatusCodes:  deniedStatusCodes,
		allowedStatusCodes: allowedStatusCodes,
	}
}

// User is expected to close the response body.
func (h *CachedClient) Do(request *http.Request) (*http.Response, error) {
	response, err := h.cacheStore.Read(request)
	if err == nil {
		return response, nil
	}

	if err != nil && !errors.Is(err, ErrNoResponse) {
		return nil, err
	}

	// Store a copy of the request body so we can retrieve it after calling
	// httpClient.Do(request).
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
	response, err = h.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	response.Request.Body = io.NopCloser(bytes.NewReader(requestBody))

	if contains(h.deniedStatusCodes, response.StatusCode) {
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

func (h *CachedClient) Get(URL string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	return h.Do(request)
}

func (h *CachedClient) Head(URL string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodHead, URL, nil)
	if err != nil {
		return nil, err
	}

	return h.Do(request)
}

func (h *CachedClient) Post(URL string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPost, URL, body)
	if err != nil {
		return nil, err
	}

	return h.Do(request)
}

func (h *CachedClient) Put(URL string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPut, URL, body)
	if err != nil {
		return nil, err
	}

	return h.Do(request)
}

func (h *CachedClient) Patch(URL string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPatch, URL, body)
	if err != nil {
		return nil, err
	}

	return h.Do(request)
}

func (h *CachedClient) Delete(URL string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodDelete, URL, body)
	if err != nil {
		return nil, err
	}

	return h.Do(request)
}

func (h *CachedClient) Options(URL string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodOptions, URL, nil)
	if err != nil {
		return nil, err
	}

	return h.Do(request)
}
