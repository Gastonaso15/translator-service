// Package config manages application configuration loaded from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration parameters for the translator service.
type Config struct {
	Port        int
	OllamaURL   string
	Model       string
	HTTPTimeout time.Duration
}

// LoadFromEnv loads configuration from environment variables with sensible defaults.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Port:        8080,
		OllamaURL:   "http://localhost:11434",
		Model:       "qwen3.5:4b",
		HTTPTimeout: 180 * time.Second,
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		portStr = strings.TrimPrefix(portStr, ":")
		p, err := strconv.Atoi(portStr)
		if err != nil || p <= 0 || p > 65535 {
			return nil, fmt.Errorf("invalid PORT value %q: must be a valid port number (1-65535)", portStr)
		}
		cfg.Port = p
	}

	if url := os.Getenv("OLLAMA_URL"); url != "" {
		cfg.OllamaURL = strings.TrimRight(url, "/")
	}

	if model := os.Getenv("LLM_MODEL"); model != "" {
		cfg.Model = model
	}

	if timeoutStr := os.Getenv("HTTP_TIMEOUT"); timeoutStr != "" {
		d, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid HTTP_TIMEOUT value %q: %w", timeoutStr, err)
		}
		cfg.HTTPTimeout = d
	}

	return cfg, nil
}

// Addr returns the network address string formatted for http.Server.
func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}
