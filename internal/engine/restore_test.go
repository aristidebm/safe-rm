package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"example.com/safe-rm/internal/config"
)

func restoreTestSetup(t *testing.T) (string, string, *config.Config) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)
	trashDir := filepath.Join(dir, "Trash")
	cfg := &config.Config{}
	cfg.TrashDir = &trashDir
	return dir, trashDir, cfg
}

func TestRestoreBasic(t *testing.T) {
	dir, _, cfg := restoreTestSetup(t)
	srcPath := filepath.Join(dir, "original.txt")
	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SoftDelete(srcPath, cfg, false); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	indexDir, _ := config.IndexDir()
	entries, _ := ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if err := Restore(entries[0], cfg, ConflictRename); err != nil {
		t.Fatalf("Restore() returned error: %v", err)
	}

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		t.Fatal("original file should exist after restore")
	}

	data, _ := os.ReadFile(srcPath)
	if string(data) != "content" {
		t.Fatalf("expected content 'content', got %q", string(data))
	}

	entries, _ = ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries after restore, got %d", len(entries))
	}
}

func TestRestoreFreeDesktop(t *testing.T) {
	dir, trashDir, cfg := restoreTestSetup(t)
	srcPath := filepath.Join(dir, "fd.txt")
	if err := os.WriteFile(srcPath, []byte("fd"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SoftDelete(srcPath, cfg, true); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	indexDir, _ := config.IndexDir()
	entries, _ := ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if !entries[0].FreeDesktop {
		t.Fatal("expected FreeDesktop=true")
	}

	if err := Restore(entries[0], cfg, ConflictRename); err != nil {
		t.Fatalf("Restore() returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(trashDir, "info", entries[0].ID+".trashinfo")); !os.IsNotExist(err) {
		t.Fatal("trashinfo should be deleted after restore")
	}
}

func TestRestoreConflictOverwrite(t *testing.T) {
	dir, _, cfg := restoreTestSetup(t)
	srcPath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(srcPath, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SoftDelete(srcPath, cfg, false); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	if err := os.WriteFile(srcPath, []byte("new file"), 0644); err != nil {
		t.Fatal(err)
	}

	indexDir, _ := config.IndexDir()
	entries, _ := ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))

	if err := Restore(entries[0], cfg, ConflictOverwrite); err != nil {
		t.Fatalf("Restore() returned error: %v", err)
	}

	data, _ := os.ReadFile(srcPath)
	if string(data) != "original" {
		t.Fatalf("expected overwritten content 'original', got %q", string(data))
	}
}

func TestRestoreConflictSkip(t *testing.T) {
	dir, _, cfg := restoreTestSetup(t)
	srcPath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(srcPath, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SoftDelete(srcPath, cfg, false); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	if err := os.WriteFile(srcPath, []byte("new file"), 0644); err != nil {
		t.Fatal(err)
	}

	indexDir, _ := config.IndexDir()
	entries, _ := ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))

	if err := Restore(entries[0], cfg, ConflictSkip); err != nil {
		t.Fatalf("Restore() returned error: %v", err)
	}

	data, _ := os.ReadFile(srcPath)
	if string(data) != "new file" {
		t.Fatalf("expected file unchanged, got %q", string(data))
	}

	entries, _ = ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry still in index, got %d", len(entries))
	}
}

func TestRestoreConflictRename(t *testing.T) {
	dir, _, cfg := restoreTestSetup(t)
	srcPath := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(srcPath, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SoftDelete(srcPath, cfg, false); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	if err := os.WriteFile(srcPath, []byte("new file"), 0644); err != nil {
		t.Fatal(err)
	}

	indexDir, _ := config.IndexDir()
	entries, _ := ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))

	if err := Restore(entries[0], cfg, ConflictRename); err != nil {
		t.Fatalf("Restore() returned error: %v", err)
	}

	restoredPath := srcPath + ".1"
	if _, err := os.Stat(restoredPath); os.IsNotExist(err) {
		t.Fatalf("expected restored file at %s", restoredPath)
	}

	data, _ := os.ReadFile(restoredPath)
	if string(data) != "original" {
		t.Fatalf("expected restored content 'original', got %q", string(data))
	}
}

func TestRestoreMissingTrashContent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	entry := &TrashEntry{
		ID:           "nonexistent",
		OriginalPath: "/tmp/fake",
		TrashedAt:    time.Now().UTC(),
	}

	cfg := &config.Config{}
	trashDir := filepath.Join(dir, "Trash")
	cfg.TrashDir = &trashDir

	err := Restore(entry, cfg, ConflictRename)
	if err == nil {
		t.Fatal("expected error for missing trash content")
	}
}