package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

// Config holds application configuration settings loaded from flags and environment variables.
type Config struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	DatabaseDSN     string
	EnableHTTPS     bool
}
type envConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
}

// NewConfig creates a new Config instance with default values from flags and environment variables.
// If parse is true, it will parse command line flags.
func NewConfig(parse bool) *Config {
	cfg := &Config{}
	var envCfg envConfig
	err := env.Parse(&envCfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base domain for short links")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "persistent storage file path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "postgres connection string")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "use https instead of http")
	if parse {
		flag.Parse()
	}

	if envCfg.ServerAddress != "" {
		cfg.ServerAddress = envCfg.ServerAddress
	}
	if envCfg.BaseURL != "" {
		cfg.BaseURL = envCfg.BaseURL
	}
	if envCfg.LogLevel != "" {
		cfg.LogLevel = envCfg.LogLevel
	}
	if envCfg.FileStoragePath != "" {
		cfg.FileStoragePath = envCfg.FileStoragePath
	}
	if envCfg.DatabaseDSN != "" {
		cfg.DatabaseDSN = envCfg.DatabaseDSN
	}
	if envCfg.EnableHTTPS {
		cfg.EnableHTTPS = envCfg.EnableHTTPS
	}
	return cfg
}
