package httpcache

import (
	"time"
)

// DefaultConfig creates a [Config] with a configuration of:
//   - Saves responses regardless of response status code, or HTTP request method;
//   - Saved responses never expire.
var DefaultConfig = NewConfigBuilder().Build()

// Config describes the configuration to use when saving to, and reading responses
// from the [Cache].
type Config struct {

	// AllowedStatusCodes describes if a HTTP response should be saved by
	// checking that it's status code is accepted by [Cache]. If the HTTP
	// response's status code is not in AllowedStatusCodes, then do not persist.
	AllowedStatusCodes *[]int

	// AllowedMethods describes if a HTTP response should be saved by checking
	// if the HTTP request's method is accepted by the [Cache]. If the HTTP
	// request's method is not in AllowedMethods, then do not persist.
	AllowedMethods *[]string

	// ExpiryTime describes when a HTTP response should be considered invalid.
	ExpiryTime *time.Duration
}
