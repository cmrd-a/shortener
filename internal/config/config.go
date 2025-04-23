package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
)

var ServerAddress string
var BaseURL string

type config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func ParseFlags() {
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	flag.StringVar(&ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&BaseURL, "b", "http://localhost:8080", "base domain for short links")
	flag.Parse()

	if cfg.ServerAddress != "" {
		ServerAddress = cfg.ServerAddress
	}
	if cfg.BaseURL != "" {
		BaseURL = cfg.BaseURL
	}
}
