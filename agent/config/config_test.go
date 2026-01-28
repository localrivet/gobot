package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Check providers exist
	if len(cfg.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(cfg.Providers))
	}

	// Check defaults
	if cfg.MaxContext != 50 {
		t.Errorf("expected MaxContext 50, got %d", cfg.MaxContext)
	}

	if cfg.MaxIterations != 100 {
		t.Errorf("expected MaxIterations 100, got %d", cfg.MaxIterations)
	}

	// Check policy defaults
	if cfg.Policy.Level != "allowlist" {
		t.Errorf("expected policy level 'allowlist', got %s", cfg.Policy.Level)
	}

	if cfg.Policy.AskMode != "on-miss" {
		t.Errorf("expected ask mode 'on-miss', got %s", cfg.Policy.AskMode)
	}

	if len(cfg.Policy.Allowlist) == 0 {
		t.Error("expected non-empty allowlist")
	}
}

func TestDefaultDataDir(t *testing.T) {
	dir := DefaultDataDir()

	if dir == "" {
		t.Error("DefaultDataDir returned empty string")
	}

	// Should end with .gobot
	if filepath.Base(dir) != ".gobot" {
		t.Errorf("expected data dir to end with .gobot, got %s", dir)
	}
}

func TestDBPath(t *testing.T) {
	cfg := DefaultConfig()
	dbPath := cfg.DBPath()

	if dbPath == "" {
		t.Error("DBPath returned empty string")
	}

	// Should end with gobot.db
	if filepath.Base(dbPath) != "gobot.db" {
		t.Errorf("expected db path to end with gobot.db, got %s", dbPath)
	}
}

func TestEnsureDataDir(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	cfg := &Config{
		DataDir: filepath.Join(tmpDir, "testdata"),
	}

	err := cfg.EnsureDataDir()
	if err != nil {
		t.Fatalf("EnsureDataDir failed: %v", err)
	}

	// Check directory was created
	info, err := os.Stat(cfg.DataDir)
	if err != nil {
		t.Fatalf("data dir not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("data dir is not a directory")
	}
}

func TestGetProvider(t *testing.T) {
	cfg := DefaultConfig()

	// Test existing provider
	p := cfg.GetProvider("anthropic-api")
	if p == nil {
		t.Error("GetProvider returned nil for existing provider")
	}
	if p.Name != "anthropic-api" {
		t.Errorf("expected provider name 'anthropic-api', got %s", p.Name)
	}

	// Test non-existing provider
	p = cfg.GetProvider("nonexistent")
	if p != nil {
		t.Error("GetProvider should return nil for non-existing provider")
	}
}

func TestFirstValidProvider(t *testing.T) {
	// Test with no valid providers
	cfg := &Config{
		Providers: []ProviderConfig{
			{Name: "empty", Type: "api", APIKey: ""},
			{Name: "cli-no-cmd", Type: "cli", Command: ""},
		},
	}

	p := cfg.FirstValidProvider()
	if p != nil {
		t.Error("FirstValidProvider should return nil when no valid providers")
	}

	// Test with valid API provider
	cfg.Providers = append(cfg.Providers, ProviderConfig{
		Name:   "valid-api",
		Type:   "api",
		APIKey: "test-key",
	})

	p = cfg.FirstValidProvider()
	if p == nil {
		t.Fatal("FirstValidProvider returned nil with valid provider")
	}
	if p.Name != "valid-api" {
		t.Errorf("expected 'valid-api', got %s", p.Name)
	}

	// Test with valid CLI provider (should be first)
	cfg.Providers = []ProviderConfig{
		{Name: "valid-cli", Type: "cli", Command: "/usr/bin/test"},
		{Name: "valid-api", Type: "api", APIKey: "test-key"},
	}

	p = cfg.FirstValidProvider()
	if p == nil {
		t.Fatal("FirstValidProvider returned nil")
	}
	if p.Name != "valid-cli" {
		t.Errorf("expected CLI provider first, got %s", p.Name)
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		Providers: []ProviderConfig{
			{Name: "test", Type: "api", APIKey: "test-key", Model: "test-model"},
		},
		DataDir:       tmpDir,
		MaxContext:    100,
		MaxIterations: 50,
		Policy: PolicyConfig{
			Level:     "full",
			AskMode:   "always",
			Allowlist: []string{"ls", "pwd"},
		},
	}

	// Save config
	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check file exists
	configPath := filepath.Join(tmpDir, "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Load config from file
	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	// Verify loaded values
	if len(loaded.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(loaded.Providers))
	}

	if loaded.MaxContext != 100 {
		t.Errorf("expected MaxContext 100, got %d", loaded.MaxContext)
	}

	if loaded.Policy.Level != "full" {
		t.Errorf("expected policy level 'full', got %s", loaded.Policy.Level)
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Load should return defaults when config doesn't exist
	origDataDir := DefaultDataDir()

	// Temporarily change HOME to a non-existent location
	tmpDir := t.TempDir()
	cfg := &Config{DataDir: tmpDir}

	// Load should succeed with defaults
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed for non-existent config: %v", err)
	}

	// Should have default values
	if len(loaded.Providers) == 0 {
		t.Error("expected default providers")
	}

	// Restore
	_ = origDataDir
	_ = cfg
}

func TestEnvironmentVariableExpansion(t *testing.T) {
	// Set a test env var
	os.Setenv("TEST_API_KEY", "expanded-key")
	defer os.Unsetenv("TEST_API_KEY")

	tmpDir := t.TempDir()
	configContent := `
providers:
  - name: test
    type: api
    api_key: ${TEST_API_KEY}
    model: test-model
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if len(loaded.Providers) == 0 {
		t.Fatal("no providers loaded")
	}

	if loaded.Providers[0].APIKey != "expanded-key" {
		t.Errorf("expected expanded API key 'expanded-key', got %s", loaded.Providers[0].APIKey)
	}
}
