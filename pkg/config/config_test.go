package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromEnv_Defaults(t *testing.T) {
	// Clear any relevant environment variables
	envVars := []string{"PORT", "OLLAMA_URL", "LLM_MODEL", "HTTP_TIMEOUT"}
	for _, k := range envVars {
		_ = os.Unsetenv(k)
	}

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error loading defaults: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected default Port 8080, got %d", cfg.Port)
	}
	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("expected default OllamaURL http://localhost:11434, got %q", cfg.OllamaURL)
	}
	if cfg.Model != "qwen3.5:4b" {
		t.Errorf("expected default Model qwen3.5:4b, got %q", cfg.Model)
	}
	if cfg.HTTPTimeout != 180*time.Second {
		t.Errorf("expected default HTTPTimeout 180s, got %v", cfg.HTTPTimeout)
	}
	if cfg.Addr() != ":8080" {
		t.Errorf("expected Addr :8080, got %q", cfg.Addr())
	}
}

func TestLoadFromEnv_Custom(t *testing.T) {
	t.Setenv("PORT", ":9090")
	t.Setenv("OLLAMA_URL", "http://ollama-service:11434/")
	t.Setenv("LLM_MODEL", "llama3.1:8b")
	t.Setenv("HTTP_TIMEOUT", "60s")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("expected Port 9090, got %d", cfg.Port)
	}
	if cfg.OllamaURL != "http://ollama-service:11434" {
		t.Errorf("expected trailing slash to be trimmed: %q", cfg.OllamaURL)
	}
	if cfg.Model != "llama3.1:8b" {
		t.Errorf("expected Model llama3.1:8b, got %q", cfg.Model)
	}
	if cfg.HTTPTimeout != 60*time.Second {
		t.Errorf("expected HTTPTimeout 60s, got %v", cfg.HTTPTimeout)
	}
	if cfg.Addr() != ":9090" {
		t.Errorf("expected Addr :9090, got %q", cfg.Addr())
	}
}

func TestLoadFromEnv_InvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-number")
	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for non-numeric port, got nil")
	}
}
