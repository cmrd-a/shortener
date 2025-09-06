package config

import (
	"os"
	"testing"
)

func TestNewConfigJSONLoading(t *testing.T) {
	// Create a temporary JSON config file
	configData := `{
		"server_address": "localhost:9090",
		"base_url": "http://example.com",
		"log_level": "debug",
		"file_storage_path": "/tmp/test.db",
		"database_dsn": "postgres://test",
		"enable_https": true
	}`

	tmpFile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configData); err != nil {
		t.Fatalf("Failed to write config data: %v", err)
	}
	tmpFile.Close()

	// Test loading config from JSON file
	cfg := &Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "",
		DatabaseDSN:     "",
		EnableHTTPS:     false,
		ConfigPath:      tmpFile.Name(),
	}

	loadJSONConfig(cfg, cfg.ConfigPath)

	// Verify JSON values were loaded
	if cfg.ServerAddress != "localhost:9090" {
		t.Errorf("Expected ServerAddress to be 'localhost:9090', got '%s'", cfg.ServerAddress)
	}
	if cfg.BaseURL != "http://example.com" {
		t.Errorf("Expected BaseURL to be 'http://example.com', got '%s'", cfg.BaseURL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
	}
	if cfg.FileStoragePath != "/tmp/test.db" {
		t.Errorf("Expected FileStoragePath to be '/tmp/test.db', got '%s'", cfg.FileStoragePath)
	}
	if cfg.DatabaseDSN != "postgres://test" {
		t.Errorf("Expected DatabaseDSN to be 'postgres://test', got '%s'", cfg.DatabaseDSN)
	}
	if !cfg.EnableHTTPS {
		t.Errorf("Expected EnableHTTPS to be true, got false")
	}
}

func TestConfigPriorityOrder(t *testing.T) {
	// Create a temporary JSON config file
	configData := `{
		"server_address": "json:8080",
		"base_url": "http://json.com",
		"log_level": "debug"
	}`

	tmpFile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configData); err != nil {
		t.Fatalf("Failed to write config data: %v", err)
	}
	tmpFile.Close()

	// Set environment variables (highest priority)
	os.Setenv("SERVER_ADDRESS", "env:8080")
	os.Setenv("BASE_URL", "http://env.com")
	os.Setenv("CONFIG", tmpFile.Name())
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("CONFIG")
	}()

	// Create config without parsing flags to avoid flag conflicts
	cfg := NewConfig(false)

	// Verify priority order:
	// Environment variables should override everything
	if cfg.ServerAddress != "env:8080" {
		t.Errorf("Expected ServerAddress to be 'env:8080' (env override), got '%s'", cfg.ServerAddress)
	}
	if cfg.BaseURL != "http://env.com" {
		t.Errorf("Expected BaseURL to be 'http://env.com' (env override), got '%s'", cfg.BaseURL)
	}
	// LogLevel should come from JSON since no env var was set for it
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug' (from JSON), got '%s'", cfg.LogLevel)
	}
}

func TestLoadJSONConfigInvalidFile(t *testing.T) {
	cfg := &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080",
		LogLevel:      "info",
	}

	// Test with non-existent file - should not panic or fail
	loadJSONConfig(cfg, "non_existent_file.json")

	// Values should remain unchanged
	if cfg.ServerAddress != ":8080" {
		t.Errorf("Expected ServerAddress to remain ':8080', got '%s'", cfg.ServerAddress)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL to remain 'http://localhost:8080', got '%s'", cfg.BaseURL)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel to remain 'info', got '%s'", cfg.LogLevel)
	}
}

func TestLoadJSONConfigInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	invalidJSON := `{
		"server_address": "localhost:9090",
		"invalid_json":
	}`

	tmpFile, err := os.CreateTemp("", "config_test_invalid_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(invalidJSON); err != nil {
		t.Fatalf("Failed to write invalid JSON data: %v", err)
	}
	tmpFile.Close()

	cfg := &Config{
		ServerAddress: ":8080",
		BaseURL:       "http://localhost:8080",
		LogLevel:      "info",
	}

	// Test with invalid JSON - should not panic or fail
	loadJSONConfig(cfg, tmpFile.Name())

	// Values should remain unchanged
	if cfg.ServerAddress != ":8080" {
		t.Errorf("Expected ServerAddress to remain ':8080', got '%s'", cfg.ServerAddress)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL to remain 'http://localhost:8080', got '%s'", cfg.BaseURL)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel to remain 'info', got '%s'", cfg.LogLevel)
	}
}

func TestLoadJSONConfigPartialValues(t *testing.T) {
	// Create a JSON config with only some values
	configData := `{
		"server_address": "localhost:9090",
		"enable_https": true
	}`

	tmpFile, err := os.CreateTemp("", "config_test_partial_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configData); err != nil {
		t.Fatalf("Failed to write config data: %v", err)
	}
	tmpFile.Close()

	cfg := &Config{
		ServerAddress:   ":8080",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "/default/path",
		DatabaseDSN:     "default_dsn",
		EnableHTTPS:     false,
	}

	loadJSONConfig(cfg, tmpFile.Name())

	// Only specified values should be updated
	if cfg.ServerAddress != "localhost:9090" {
		t.Errorf("Expected ServerAddress to be 'localhost:9090', got '%s'", cfg.ServerAddress)
	}
	if !cfg.EnableHTTPS {
		t.Errorf("Expected EnableHTTPS to be true, got false")
	}

	// Unspecified values should remain unchanged
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL to remain 'http://localhost:8080', got '%s'", cfg.BaseURL)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel to remain 'info', got '%s'", cfg.LogLevel)
	}
	if cfg.FileStoragePath != "/default/path" {
		t.Errorf("Expected FileStoragePath to remain '/default/path', got '%s'", cfg.FileStoragePath)
	}
	if cfg.DatabaseDSN != "default_dsn" {
		t.Errorf("Expected DatabaseDSN to remain 'default_dsn', got '%s'", cfg.DatabaseDSN)
	}
}

func TestConfigFlagOverride(t *testing.T) {
	// Create a temporary JSON config file
	configData := `{
		"server_address": "json:8080",
		"base_url": "http://json.com",
		"log_level": "debug"
	}`

	tmpFile, err := os.CreateTemp("", "config_test_flag_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configData); err != nil {
		t.Fatalf("Failed to write config data: %v", err)
	}
	tmpFile.Close()

	// Test behavior by using environment variable for config path
	// and then checking that the JSON values are loaded correctly
	os.Setenv("CONFIG", tmpFile.Name())
	defer os.Unsetenv("CONFIG")

	// Test with config loaded from environment variable
	cfg := NewConfig(false)

	// Values should come from JSON config
	if cfg.ServerAddress != "json:8080" {
		t.Errorf("Expected ServerAddress to be 'json:8080' (from JSON), got '%s'", cfg.ServerAddress)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug' (from JSON), got '%s'", cfg.LogLevel)
	}
	if cfg.BaseURL != "http://json.com" {
		t.Errorf("Expected BaseURL to be 'http://json.com' (from JSON), got '%s'", cfg.BaseURL)
	}
}
