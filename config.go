package httpcache

import "net/http"

// Add doc
type Config struct {

	// Add doc
	Cache Cache

	// Add doc
	AllowedStatusCodes *[]int

	// Add doc
	AllowedMethods *[]string
}

// Add doc
func NewConfig() *Config {
	return &Config{}
}

// Add doc
func (c *Config) WithAllowedStatusCodes(allowedStatusCodes []int) *Config {
	c.AllowedStatusCodes = &allowedStatusCodes
	return c
}

// Add doc
func (c *Config) WithAllowedMethods(allowedMethods []string) *Config {
	c.AllowedMethods = &allowedMethods
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
		WithAllowedStatusCodes([]int{
			http.StatusOK,
		}).
		WithAllowedMethods([]string{
			http.MethodGet,
		}).
		WithCache(cache)

	return defaultConfig, nil
}
