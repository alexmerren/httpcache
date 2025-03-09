package httpcache

import "time"

// If adding new fields to [Config], then add the corresponding field and
// methods to [configBuilder]. Configuration values in [configBuilder] always
// use pointer types.  This allows the [Config] decide which fields are
// required.

// configBuilder is an internal structure to create configs with required
// parameters.
type configBuilder struct {

	// allowedStatusCodes only persists HTTP responses that have an appropriate
	// status code (i.e. 200).
	//
	// This is a required field.
	allowedStatusCodes *[]int

	// allowedMethods only persists HTTP responses that use an appropriate HTTP
	// method (i.e. "GET").
	//
	// This is a required field.
	allowedMethods *[]string

	// expiryTime invalidates HTTP responses after a duration has elapsed from
	// [time.Now]. Set to nil for no expiry.
	expiryTime *time.Duration
}

func NewConfigBuilder() *configBuilder {
	return &configBuilder{}
}

func (c *configBuilder) WithAllowedStatusCodes(allowedStatusCodes []int) *configBuilder {
	c.allowedStatusCodes = &allowedStatusCodes
	return c
}

func (c *configBuilder) WithAllowedMethods(allowedMethods []string) *configBuilder {
	c.allowedMethods = &allowedMethods
	return c
}

func (c *configBuilder) WithExpiryTime(expiryDuration time.Duration) *configBuilder {
	c.expiryTime = &expiryDuration
	return c
}

// Build constructs a [Config] that is ready to be consumed by [Transport]. If
// the configuration passed by [configBuilder] is invalid, it will panic.
func (c *configBuilder) Build() *Config {
	return &Config{
		AllowedStatusCodes: *c.allowedStatusCodes,
		AllowedMethods:     *c.allowedMethods,
		ExpiryTime:         c.expiryTime,
	}
}
