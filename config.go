package httpcache

import (
	"net/http"
	"time"
)

var (
	defaultAllowedStatusCodes = []int{http.StatusOK}
	defaultAllowedMethods     = []string{http.MethodGet}
	defaultExpiryTime         = time.Duration(60*24*7) * time.Minute
)

// Config describes the configuration to use when saving and reading responses
// from [Cache] using the [Transport].
type Config struct {

	// AllowedStatusCodes describes if a HTTP response should be saved by
	// checking that it's status code is accepted by [Cache]. If the HTTP
	// response's status code is not in AllowedStatusCodes, then do not persist.
	//
	// This is a required field.
	AllowedStatusCodes []int

	// AllowedMethods describes if a HTTP response should be saved by checking
	// if the HTTP request's method is accepted by the [Cache]. If the HTTP
	// request's method is not in AllowedMethods, then do not persist.
	//
	// This is a required field.
	AllowedMethods []string

	// ExpiryTime describes when a HTTP response should be considered invalid.
	ExpiryTime *time.Duration
}

// DefaultConfig creates a [Config] with default values, namely:
//   - AllowedStatusCodes: [http.StatusOK]
//   - AllowedMethods: [http.MethodGet]
//   - ExpiryTime: 7 days.
func DefaultConfig() (*Config, error) {
	defaultConfig := NewConfigBuilder().
		WithAllowedStatusCodes(defaultAllowedStatusCodes).
		WithAllowedMethods(defaultAllowedMethods).
		WithExpiryTime(defaultExpiryTime).
		Build()

	return defaultConfig, nil
}
