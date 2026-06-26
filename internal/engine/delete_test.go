package engine

import (
	"os"
	"path/filepath"
	"testing"

	"example.com/safe-rm/internal/config"
)

func TestPolicyRoutingSoftDelete(t *testing.T) {
	cfg := &config.Config{}
	policy, err := Route("/some/file.txt", cfg)
	if err != nil {
		t.Fatalf("Route() returned error: %v", err)
	}
	if policy != PolicySoftDelete {
		t.Fatalf("expected PolicySoftDelete, got %v", policy)
	}
}

func TestPolicyRoutingBypass(t *testing.T) {
	cfg := &config.Config{
		BypassList: []string{"*.tmp"},
		DangerList: []string{},
	}
	policy, err := Route("/some/file.tmp", cfg)
	if err != nil {
		t.Fatalf("Route() returned error: %v", err)
	}
	if policy != PolicyPermanent {
		t.Fatalf("expected PolicyPermanent, got %v", policy)
	}
}

func TestPolicyRoutingDanger(t *testing.T) {
	cfg := &config.Config{
		DangerList: []string{"*.secret"},
	}
	policy, err := Route("/some/password.secret", cfg)
	if err != nil {
		t.Fatalf("Route() returned error: %v", err)
	}
	if policy != PolicyDanger {
		t.Fatalf("expected PolicyDanger, got %v", policy)
	}
}

func TestPolicyRoutingDangerPermanent(t *testing.T) {
	cfg := &config.Config{
		DangerList: []string{"*.secret"},
		BypassList: []string{"*.secret"},
	}
	policy, err := Route("/some/password.secret", cfg)
	if err != nil {
		t.Fatalf("Route() returned error: %v", err)
	}
	if policy != PolicyDangerPermanent {
		t.Fatalf("expected PolicyDangerPermanent, got %v", policy)
	}
}

func TestSoftDeleteFile(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "testfile.txt")
	if err := os.WriteFile(srcPath, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	trashDir := filepath.Join(dir, "Trash")
	cfg := &config.Config{}
	cfg.TrashDir = &trashDir

	if err := SoftDelete(srcPath, cfg, false); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Fatal("source file should no longer exist")
	}

	entries, err := os.ReadDir(filepath.Join(trashDir, "files"))
	if err != nil {
		t.Fatalf("failed to read trash files dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file in trash, got %d", len(entries))
	}

	if _, err := os.Stat(filepath.Join(trashDir, "info")); !os.IsNotExist(err) {
		t.Fatal("non-FreeDesktop mode should not create info dir")
	}
}

func TestSoftDeleteFreeDesktop(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "testfile.txt")
	if err := os.WriteFile(srcPath, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	trashDir := filepath.Join(dir, "Trash")
	cfg := &config.Config{}
	cfg.TrashDir = &trashDir

	if err := SoftDelete(srcPath, cfg, true); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	files, err := os.ReadDir(filepath.Join(trashDir, "files"))
	if err != nil {
		t.Fatalf("failed to read trash files dir: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file in trash, got %d", len(files))
	}

	infos, err := os.ReadDir(filepath.Join(trashDir, "info"))
	if err != nil {
		t.Fatalf("failed to read trash info dir: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 info file, got %d", len(infos))
	}
}

func TestSoftDeleteDirectory(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "mydir")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "inner.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	trashDir := filepath.Join(dir, "Trash")
	cfg := &config.Config{}
	cfg.TrashDir = &trashDir

	if err := SoftDelete(srcDir, cfg, false); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Fatal("source directory should no longer exist")
	}

	entries, err := os.ReadDir(filepath.Join(trashDir, "files"))
	if err != nil {
		t.Fatalf("failed to read trash files dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in trash, got %d", len(entries))
	}
}

func TestPermanentDelete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todelete.txt")
	if err := os.WriteFile(path, []byte("bye"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := PermanentDelete(path); err != nil {
		t.Fatalf("PermanentDelete() returned error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("file should be permanently deleted")
	}
}

func TestIsRootOrCWD(t *testing.T) {
	if !IsRootOrCWD("/") {
		t.Fatal("expected IsRootOrCWD('/') to be true")
	}

	if IsRootOrCWD("/home") {
		t.Fatal("expected IsRootOrCWD('/home') to be false")
	}

	cwd, _ := os.Getwd()
	if !IsRootOrCWD(cwd) {
		t.Fatal("expected IsRootOrCWD(cwd) to be true")
	}
}

func TestSoftDeleteWithJSONLEntry(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(srcPath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	trashDir := filepath.Join(dir, "Trash")
	cfg := &config.Config{}
	cfg.TrashDir = &trashDir

	t.Setenv("XDG_DATA_HOME", dir)
	if err := SoftDelete(srcPath, cfg, false); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	indexDir, _ := config.IndexDir()
	entries, err := ReadAllEntries(filepath.Join(indexDir, "trash.jsonl"))
	if err != nil {
		t.Fatalf("ReadAllEntries() returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in JSONL, got %d", len(entries))
	}

	if entries[0].OriginalPath != srcPath {
		t.Fatalf("expected OriginalPath=%q, got %q", srcPath, entries[0].OriginalPath)
	}
}

func TestSoftDeleteDisappearedFile(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "ghost.txt")

	trashDir := filepath.Join(dir, "Trash")
	cfg := &config.Config{}
	cfg.TrashDir = &trashDir

	err := SoftDelete(srcPath, cfg, false)
	if err != nil {
		t.Fatalf("SoftDelete() on non-existent file returned error: %v", err)
	}
}
