package config

import "github.com/caarlos0/env/v6"

// Config stores server configuration. Config value can be taken from environment variables or hardcoded.
type Config struct {
	ServerPort    int    `env:"SERVER_PORT" envDefault:"false"`
	ServerAddress string `env:"SERVER_ADDRESS"`
	DebugMode     bool   `env:"DEBUG_MODE" envDefault:"false"`
}

// ParseEnvVars parses environment variables into config.Config.
func (cfg *Config) ParseEnvVars() error {
	return env.Parse(cfg)
}
