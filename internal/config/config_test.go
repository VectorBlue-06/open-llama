package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Generation.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", cfg.Generation.Temperature)
	}
	if cfg.Template.Default != "chatml" {
		t.Errorf("expected template chatml, got %s", cfg.Template.Default)
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := Defaults()
	cfg.Generation.Temperature = 0.5
	cfg.Model.Default = "test-model.gguf"

	if err := SaveTo(cfg, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Generation.Temperature != 0.5 {
		t.Errorf("expected temperature 0.5, got %f", loaded.Generation.Temperature)
	}
	if loaded.Model.Default != "test-model.gguf" {
		t.Errorf("expected model test-model.gguf, got %s", loaded.Model.Default)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("expected defaults, got version %d", cfg.Version)
	}
	// File should have been created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("{invalid json"), 0600)

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("expected no error for invalid JSON, got %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("expected defaults on invalid JSON, got version %d", cfg.Version)
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expanded := ExpandPath("~/test/path")
	expected := filepath.Join(home, "test/path")
	if expanded != expected {
		t.Errorf("expected %s, got %s", expected, expanded)
	}

	// Non-home path should stay the same
	plain := ExpandPath("/absolute/path")
	if plain != "/absolute/path" {
		t.Errorf("expected /absolute/path, got %s", plain)
	}
}
