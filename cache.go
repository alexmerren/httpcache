package httpcache

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Cache is the entrypoint for saving and reading responses. This can be
// implemented for a custom method to cache responses.
type Cache interface {

	// Save a response for a HTTP request using [context.Background].
	Save(response *http.Response) error

	// Read a saved response for a HTTP request using [context.Background].
	Read(request *http.Request) (*ReadResult, error)

	// Delete a saved response using [context.Background].
	Delete(response *http.Response) error

	// Save a response for a HTTP request with a [context.Context].
	SaveContext(ctx context.Context, response *http.Response) error

	// Read a saved response for a HTTP request with a [context.Context]. If no
	// response is saved for the corresponding request.
	ReadContext(ctx context.Context, request *http.Request) (*ReadResult, error)

	// Delete a saved response with a [context.Context].
	DeleteContext(ctx context.Context, response *http.Response) error
}

// ReadResult is the result of a successful read operation on the cache.
type ReadResult struct {
	response  *http.Response
	createdAt *time.Time
}

var (
	// ErrNoResult describes when the cache does not have a response stored.
	// [Transport] will check if ErrNoResponse is returned from [Cache.Read]. If
	// ErrNoResponse is returned, then the request/response will be saved with [Save].
	ErrNoResult = errors.New("found no result from read")
)
