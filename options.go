package httpcache

func WithName(name string) func(*CachedRoundTripper) {
	return func(c *CachedRoundTripper) {
		c.cacheStore = NewSqliteResponseStore(name)
	}
}

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

func WithCacheStore(store ResponseStorer) func(*CachedRoundTripper) {
	return func(c *CachedRoundTripper) {
		c.cacheStore = store
	}
}
