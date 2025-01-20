package httpcache

import (
	"context"
	"errors"
	"net/http"
)

// Add doc
type Cache interface {

	// Add doc
	Save(response *http.Response) error

	// Add doc
	Read(request *http.Request) (*http.Response, error)

	// Add doc
	SaveContext(ctx context.Context, response *http.Response) error

	// Add doc
	ReadContext(ctx context.Context, request *http.Request) (*http.Response, error)
}

var (
	// Add doc
	ErrNoResponse = errors.New("no stored response")
)
