package config

import (
	"time"

	"github.com/caarlos0/env/v6"
)

// Config stores server configuration. Config value can be taken from environment variables or hardcoded.
type Config struct {
	ProxyServerHost            string `env:"PROXY_SERVER_HOST"`
	ProxyServerPort            int    `env:"PROXY_SERVER_PORT"`
	ProxyServerAddress         string `env:"PROXY_SERVER_ADDRESS"`
	ProxyUserAgent             string `env:"PROXY_USER_AGENT"`
	AccountUsername            string `env:"AHREFS_ACCOUNT_USERNAME"`
	AccountPassword            string `env:"AHREFS_ACCOUNT_PASSWORD"`
	CacheRedisAddress          string `env:"CACHE_REDIS_ADDRESS"`
	CacheRedisPassword         string `env:"CACHE_REDIS_PASSWORD"`
	ProxyDebugMode             bool   `env:"PROXY_DEBUG_MODE" envDefault:"false"`
	ProxyServerShutdownTimeout time.Duration
}

func New() (*Config, error) {
	cfg := &Config{
		ProxyServerShutdownTimeout: 5 * time.Second,
	}
	if err := cfg.ParseEnvVars(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ParseEnvVars parses environment variables into config.Config.
func (cfg *Config) ParseEnvVars() error {
	return env.Parse(cfg)
}
