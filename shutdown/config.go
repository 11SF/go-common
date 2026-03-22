package shutdown

import (
	"errors"
	"time"
)

const DefaultTimeout = 5 * time.Second

// Config holds configuration for graceful shutdown.
type Config struct {
	Addr    string        // required — listen address e.g. ":8080"
	Timeout time.Duration // optional — shutdown timeout, defaults to DefaultTimeout (5s)
}

func (c *Config) validate() error {
	if c.Addr == "" {
		return errors.New("shutdown: addr is required")
	}
	return nil
}

func (c *Config) withDefaults() {
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
}
