package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

var ServerAddress string
var BaseURL string
var LogLevel string
var FileStoragePath string

type config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func ParseFlags() {
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&BaseURL, "b", "http://localhost:8080", "base domain for short links")
	flag.StringVar(&LogLevel, "l", "info", "log level")
	flag.StringVar(&FileStoragePath, "f", "storage.txt", "persistent storage file path")
	flag.Parse()

	if cfg.ServerAddress != "" {
		ServerAddress = cfg.ServerAddress
	}
	if cfg.BaseURL != "" {
		BaseURL = cfg.BaseURL
	}
	if cfg.LogLevel != "" {
		LogLevel = cfg.LogLevel
	}
	if cfg.FileStoragePath != "" {
		FileStoragePath = cfg.FileStoragePath
	}
}
