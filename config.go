package httpcache

import "net/http"

// Add doc
type Config struct {

	// Add doc
	Cache Cache

	// Add doc
	DeniedStatusCodes *[]int

	// Add doc
	AllowedStatusCodes *[]int
}

// Add doc
func NewConfig() *Config {
	return &Config{}
}

// Add doc
func (c *Config) WithDeniedStatusCodes(deniedStatusCodes []int) *Config {
	c.DeniedStatusCodes = &deniedStatusCodes
	return c
}

// Add doc
func (c *Config) WithAllowedStatusCodes(allowedStatusCodes []int) *Config {
	c.AllowedStatusCodes = &allowedStatusCodes
	return c
}

// Add doc
func (c *Config) WithCache(cache Cache) *Config {
	c.Cache = cache
	return c
}

func DefaultConfig() (*Config, error) {
	cache, err := NewSqliteCache("httpcache.sqlite")
	if err != nil {
		return nil, err
	}

	defaultConfig := NewConfig().
		WithDeniedStatusCodes([]int{
			http.StatusNotFound,
			http.StatusBadRequest,
			http.StatusForbidden,
			http.StatusUnauthorized,
			http.StatusMethodNotAllowed,
		}).
		WithAllowedStatusCodes([]int{
			http.StatusOK,
		}).
		WithCache(cache)

	return defaultConfig, nil
}
