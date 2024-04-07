package httpcache

import "time"

func WithDeniedStatusCodes(deniedStatusCodes []int) func(*CachedRoundTripper) {
	return func(c *CachedRoundTripper) {
		c.deniedStatusCodes = deniedStatusCodes
	}
}

func WithAllowedStatusCodes(allowedStatusCodes []int) func(*CachedRoundTripper) {
	return func(c *CachedRoundTripper) {
		c.allowedStatusCodes = allowedStatusCodes
	}
}

func WithExpiryTime(expiryTime time.Duration) func(*CachedRoundTripper) {
	return func(c *CachedRoundTripper) {
		c.expiryTime = expiryTime
	}
}

func WithName(name string) func(*CachedRoundTripper) {
	return func(c *CachedRoundTripper) {
		c.store = newSqliteResponseStore(name)
	}
}

func WithCacheStore(store ResponseStorer) func(*CachedRoundTripper) {
	return func(c *CachedRoundTripper) {
		c.store = store
	}
}
