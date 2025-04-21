package main

import (
	"flag"
	"github.com/caarlos0/env/v11"
)

var serverAddress string
var baseURL string

type config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func parseFlags() {
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	flag.StringVar(&serverAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&baseURL, "b", "http://localhost:8080", "base domain for short links")
	flag.Parse()

	if cfg.ServerAddress != "" {
		serverAddress = cfg.ServerAddress
	}
	if cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}
}
