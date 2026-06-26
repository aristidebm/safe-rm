package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNoConfigFile(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
}

func TestLoadWithSAFE_RM_CONFIG(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.toml")
	if err := os.WriteFile(path, []byte(`trash_dir = "/custom/trash"`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SAFE_RM_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.TrashDir == nil || *cfg.TrashDir != "/custom/trash" {
		t.Fatalf("expected TrashDir=/custom/trash, got %v", cfg.TrashDir)
	}
}

func TestLoadWithXDGConfig(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "safe-rm")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(path, []byte("bypass_list = [\"*.tmp\"]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("SAFE_RM_CONFIG", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if len(cfg.BypassList) != 1 || cfg.BypassList[0] != "*.tmp" {
		t.Fatalf("expected BypassList=[\"*.tmp\"], got %v", cfg.BypassList)
	}
}

func TestResolvedTrashDirCustom(t *testing.T) {
	trashDir := "/custom/trash"
	cfg := &Config{TrashDir: &trashDir}

	resolved, err := cfg.ResolvedTrashDir()
	if err != nil {
		t.Fatalf("ResolvedTrashDir() returned error: %v", err)
	}

	if resolved != "/custom/trash" {
		t.Fatalf("expected /custom/trash, got %s", resolved)
	}
}

func TestResolvedTrashDirDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	cfg := &Config{}

	resolved, err := cfg.ResolvedTrashDir()
	if err != nil {
		t.Fatalf("ResolvedTrashDir() returned error: %v", err)
	}

	expected := filepath.Join(dir, "Trash")
	if resolved != expected {
		t.Fatalf("expected %s, got %s", expected, resolved)
	}
}

func TestIndexDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	got, err := IndexDir()
	if err != nil {
		t.Fatalf("IndexDir() returned error: %v", err)
	}

	expected := filepath.Join(dir, "safe-rm")
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestExpandPathTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	got := expandPath("~/test")
	expected := filepath.Join(home, "test")
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestExpandPathEnvVar(t *testing.T) {
	t.Setenv("MY_TRASH", "/some/path")

	got := expandPath("$MY_TRASH/subdir")
	expected := "/some/path/subdir"
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("garbage [[[invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SAFE_RM_CONFIG", path)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}
