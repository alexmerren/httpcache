package httpcache

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

var DefaultClient = NewDefaultClient()

var unacceptableResponseCodes = []int{
	http.StatusNotFound,
	http.StatusBadRequest,
	http.StatusForbidden,
	http.StatusUnauthorized,
	http.StatusMethodNotAllowed,
}

type CachedClient struct {
	httpClient *http.Client
	cacheStore ResponseStorer
}

func NewDefaultClient() *CachedClient {
	return &CachedClient{
		httpClient: http.DefaultClient,
		cacheStore: NewDefaultResponseStore(),
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

	// Reset the request body so that it can be read by the cache store.
	response, err = h.httpClient.Do(request)
	if err != nil {
		request.Body.Close()
		return nil, err
	}
	response.Request.Body = io.NopCloser(bytes.NewReader(requestBody))

	// Not sure what to do here... maybe set cachable response codes in struct?
	if contains(unacceptableResponseCodes, response.StatusCode) {
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
