package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_EnvVarTakesPriority(t *testing.T) {
	// Set up env var
	os.Setenv("TIBBER_TOKEN", "env-token-123")
	defer os.Unsetenv("TIBBER_TOKEN")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Token != "env-token-123" {
		t.Errorf("Token = %q, want %q", cfg.Token, "env-token-123")
	}
}

func TestLoad_DefaultFormat(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Format != "pretty" {
		t.Errorf("Format = %q, want %q", cfg.Format, "pretty")
	}
}

func TestLoad_ConfigFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `token: "file-token-456"
home_id: "home-123"
format: "json"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Ensure no env var interference
	os.Unsetenv("TIBBER_TOKEN")
	os.Unsetenv("TIBBER_HOME_ID")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Token != "file-token-456" {
		t.Errorf("Token = %q, want %q", cfg.Token, "file-token-456")
	}
	if cfg.HomeID != "home-123" {
		t.Errorf("HomeID = %q, want %q", cfg.HomeID, "home-123")
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q, want %q", cfg.Format, "json")
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `token: "file-token"
home_id: "file-home"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set env var - should override file
	os.Setenv("TIBBER_TOKEN", "env-token")
	defer os.Unsetenv("TIBBER_TOKEN")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Env should win
	if cfg.Token != "env-token" {
		t.Errorf("Token = %q, want %q (env should override file)", cfg.Token, "env-token")
	}
	// File value should still be used for home_id
	if cfg.HomeID != "file-home" {
		t.Errorf("HomeID = %q, want %q", cfg.HomeID, "file-home")
	}
}

func TestValidate_MissingToken(t *testing.T) {
	cfg := &Config{}
	err := cfg.Validate()

	if err == nil {
		t.Error("Validate() should return error for missing token")
	}
}

func TestValidate_WithToken(t *testing.T) {
	cfg := &Config{Token: "some-token"}
	err := cfg.Validate()

	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()

	if path == "" {
		t.Error("DefaultConfigPath() returned empty string")
	}

	if !filepath.IsAbs(path) {
		t.Errorf("DefaultConfigPath() = %q, want absolute path", path)
	}

	if filepath.Base(path) != "config.yaml" {
		t.Errorf("DefaultConfigPath() basename = %q, want config.yaml", filepath.Base(path))
	}
}
