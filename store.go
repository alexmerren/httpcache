package httpcache

import (
	"context"
	"errors"
	"net/http"
)

type ResponseStorer interface {
	Save(response *http.Response) error
	Read(request *http.Request) (*http.Response, error)
	SaveContext(ctx context.Context, response *http.Response) error
	ReadContext(ctx context.Context, request *http.Request) (*http.Response, error)
}

var (
	ErrNoResponse = errors.New("no stored response")
)
