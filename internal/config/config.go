package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
)

// Config holds application configuration settings loaded from various sources with the following priority order:
// 1. Default values (lowest priority)
// 2. JSON configuration file (overrides defaults)
// 3. Command line flags (overrides JSON)
// 4. Environment variables (highest priority, overrides everything)
//
// JSON configuration file can be specified via:
// - Command line flag: -c config.json
// - Environment variable: CONFIG=config.json
//
// Example JSON configuration file:
//
//	{
//	  "server_address": "localhost:8080",
//	  "base_url": "http://localhost",
//	  "log_level": "debug",
//	  "file_storage_path": "/path/to/storage.db",
//	  "database_dsn": "postgres://user:pass@host/db",
//	  "enable_https": true
//	}
type Config struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	DatabaseDSN     string
	EnableHTTPS     bool
	ConfigPath      string
}

type envJSONConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL         string `env:"BASE_URL" json:"base_url"`
	LogLevel        string `env:"LOG_LEVEL" json:"log_level"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigPath      string `env:"CONFIG"`
}

// NewConfig creates a new Config instance with configuration loaded in priority order.
// See Config struct documentation for detailed information about priority and usage.
// If parse is true, it will parse command line flags.
func NewConfig(parse bool) *Config {
	// Step 1: Start with default values
	cfg := &Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "",
		DatabaseDSN:     "",
		EnableHTTPS:     false,
		ConfigPath:      "",
	}

	// Step 2: Parse environment variables to get config path
	var envCfg envJSONConfig
	err := env.Parse(&envCfg)
	if err != nil {
		log.Fatal(err)
	}

	// Use environment config path if available
	if envCfg.ConfigPath != "" {
		cfg.ConfigPath = envCfg.ConfigPath
	}

	// Step 3: Parse command line flags (they will temporarily hold flag values)
	var flagValues *Config
	if parse && !flag.Parsed() {
		flagValues = &Config{}
		flag.StringVar(&flagValues.ServerAddress, "a", cfg.ServerAddress, "address and port to run server")
		flag.StringVar(&flagValues.BaseURL, "b", cfg.BaseURL, "base domain for short links")
		flag.StringVar(&flagValues.LogLevel, "l", cfg.LogLevel, "log level")
		flag.StringVar(&flagValues.FileStoragePath, "f", cfg.FileStoragePath, "persistent storage file path")
		flag.StringVar(&flagValues.DatabaseDSN, "d", cfg.DatabaseDSN, "postgres connection string")
		flag.BoolVar(&flagValues.EnableHTTPS, "s", cfg.EnableHTTPS, "use https instead of http")
		flag.StringVar(&flagValues.ConfigPath, "c", cfg.ConfigPath, "config file path")

		flag.Parse()

		// Update config path from flag if provided
		if flagValues.ConfigPath != cfg.ConfigPath {
			cfg.ConfigPath = flagValues.ConfigPath
		}
	}

	// Step 4: Load JSON config (overrides defaults)
	if cfg.ConfigPath != "" {
		loadJSONConfig(cfg, cfg.ConfigPath)
	}

	// Step 5: Apply flag values (override JSON) - only if they were explicitly set
	if flagValues != nil {
		explicitFlags := make(map[string]bool)
		flag.Visit(func(f *flag.Flag) {
			explicitFlags[f.Name] = true
		})

		if explicitFlags["a"] {
			cfg.ServerAddress = flagValues.ServerAddress
		}
		if explicitFlags["b"] {
			cfg.BaseURL = flagValues.BaseURL
		}
		if explicitFlags["l"] {
			cfg.LogLevel = flagValues.LogLevel
		}
		if explicitFlags["f"] {
			cfg.FileStoragePath = flagValues.FileStoragePath
		}
		if explicitFlags["d"] {
			cfg.DatabaseDSN = flagValues.DatabaseDSN
		}
		if explicitFlags["s"] {
			cfg.EnableHTTPS = flagValues.EnableHTTPS
		}
		if explicitFlags["c"] && flagValues.ConfigPath != cfg.ConfigPath {
			cfg.ConfigPath = flagValues.ConfigPath
			// Reload JSON config if path was changed via flag
			loadJSONConfig(cfg, cfg.ConfigPath)
		}
	}

	// Step 6: Apply environment variables (highest priority, override everything)
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

// loadJSONConfig loads configuration from a JSON file
func loadJSONConfig(cfg *Config, configPath string) {
	f, err := os.OpenFile(configPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Printf("Warning: failed to open config file %s: %v", configPath, err)
		return
	}
	defer f.Close()

	var jsonCfg envJSONConfig
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&jsonCfg)
	if err != nil {
		log.Printf("Warning: failed to parse JSON config file %s: %v", configPath, err)
		return
	}

	// Apply JSON config values if they are not empty
	if jsonCfg.ServerAddress != "" {
		cfg.ServerAddress = jsonCfg.ServerAddress
	}
	if jsonCfg.BaseURL != "" {
		cfg.BaseURL = jsonCfg.BaseURL
	}
	if jsonCfg.LogLevel != "" {
		cfg.LogLevel = jsonCfg.LogLevel
	}
	if jsonCfg.FileStoragePath != "" {
		cfg.FileStoragePath = jsonCfg.FileStoragePath
	}
	if jsonCfg.DatabaseDSN != "" {
		cfg.DatabaseDSN = jsonCfg.DatabaseDSN
	}
	// For boolean values, we apply them directly since false is a valid value
	cfg.EnableHTTPS = jsonCfg.EnableHTTPS
}
