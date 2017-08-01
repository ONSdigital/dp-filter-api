package config

import "github.com/ian-kent/gofigure"

// Config is the filing resource handler config
type Config struct {
	BindAddr    string `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"The port to bind to"`
	Host        string `env:"HOST" flag:"host" flagDesc:"The host name used to build URLs"`
	PostgresURL string `env:"POSTGRES_URL" flag:"postgres-url" flagDesc:"The URL address to connect to a postgres instance'"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:    ":22100",
		Host:        "http://localhost:22100",
		PostgresURL: "user=dp dbname=FilterJobs sslmode=disable",
	}

	if err := gofigure.Gofigure(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
