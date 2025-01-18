package httpcache

import (
	"time"
)

func WithDeniedStatusCodes(deniedStatusCodes []int) func(*CachedRoundTripper) error {
	return func(c *CachedRoundTripper) error {
		c.deniedStatusCodes = deniedStatusCodes
		return nil
	}
}

func WithAllowedStatusCodes(allowedStatusCodes []int) func(*CachedRoundTripper) error {
	return func(c *CachedRoundTripper) error {
		c.allowedStatusCodes = allowedStatusCodes
		return nil
	}
}

func WithExpiryTime(expiryTime time.Duration) func(*CachedRoundTripper) error {
	return func(c *CachedRoundTripper) error {
		c.expiryTime = expiryTime
		return nil
	}
}

func WithName(name string) func(*CachedRoundTripper) error {
	return func(c *CachedRoundTripper) error {
		sqliteStore, err := newSqliteResponseStore(name)
		if err != nil {
			return err
		}
		c.store = sqliteStore

		return nil
	}
}

func WithCacheStore(store ResponseStorer) func(*CachedRoundTripper) error {
	return func(c *CachedRoundTripper) error {
		c.store = store
		return nil
	}
}
