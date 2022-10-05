package config

import "github.com/caarlos0/env/v6"

// Config stores server configuration. Config value can be taken from environment variables or hardcoded.
type Config struct {
	ServerPort      int    `env:"SERVER_PORT" envDefault:"false"`
	ServerAddress   string `env:"SERVER_ADDRESS"`
	AccountUsername string `env:"AHREFS_ACCOUNT_USERNAME"`
	AccountPassword string `env:"AHREFS_ACCOUNT_PASSWORD"`
	DebugMode       bool   `env:"DEBUG_MODE" envDefault:"false"`
	UserAgent       string
}

func New() (*Config, error) {
	cfg := &Config{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36",
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
