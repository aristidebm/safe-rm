package engine

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewEntryForFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	entry, err := NewEntry(path, false)
	if err != nil {
		t.Fatalf("NewEntry() returned error: %v", err)
	}

	if entry.ID == "" || len(entry.ID) != 12 {
		t.Fatalf("expected 12-char hex ID, got %q (len=%d)", entry.ID, len(entry.ID))
	}

	if entry.OriginalPath != path {
		t.Fatalf("expected OriginalPath=%q, got %q", path, entry.OriginalPath)
	}

	if entry.TrashedAt.IsZero() {
		t.Fatal("expected TrashedAt to be set")
	}

	if entry.Size != 5 {
		t.Fatalf("expected Size=5, got %d", entry.Size)
	}

	if entry.IsDir {
		t.Fatal("expected IsDir=false")
	}

	if entry.Checksum == "" {
		t.Fatal("expected non-empty Checksum")
	}

	if entry.Permanent {
		t.Fatal("expected Permanent=false")
	}

	if entry.FreeDesktop {
		t.Fatal("expected FreeDesktop=false")
	}
}

func TestNewEntryForDir(t *testing.T) {
	dir := t.TempDir()

	entry, err := NewEntry(dir, true)
	if err != nil {
		t.Fatalf("NewEntry() returned error: %v", err)
	}

	if entry.Checksum != "" {
		t.Fatalf("expected empty Checksum for dir, got %q", entry.Checksum)
	}

	if !entry.IsDir {
		t.Fatal("expected IsDir=true")
	}

	if entry.Size != 0 {
		t.Fatalf("expected Size=0 for dir, got %d", entry.Size)
	}
}

func TestAppendAndReadEntries(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "trash.jsonl")

	e1 := &TrashEntry{
		ID:           "abc123",
		OriginalPath: "/tmp/file1.txt",
		TrashedAt:    mustParseTime("2026-06-26T14:32:00Z"),
		Size:         100,
		IsDir:        false,
		Checksum:     "checksum1",
	}

	e2 := &TrashEntry{
		ID:           "def456",
		OriginalPath: "/tmp/file2.txt",
		TrashedAt:    mustParseTime("2026-06-26T15:01:12Z"),
		Size:         200,
		IsDir:        false,
		Checksum:     "checksum2",
	}

	if err := AppendEntry(indexPath, e1); err != nil {
		t.Fatalf("AppendEntry() returned error: %v", err)
	}
	if err := AppendEntry(indexPath, e2); err != nil {
		t.Fatalf("AppendEntry() returned error: %v", err)
	}

	entries, err := ReadAllEntries(indexPath)
	if err != nil {
		t.Fatalf("ReadAllEntries() returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].ID != "abc123" || entries[1].ID != "def456" {
		t.Fatalf("unexpected ordering: %v, %v", entries[0].ID, entries[1].ID)
	}
}

func TestWriteAllEntries(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "trash.jsonl")

	entries := []*TrashEntry{
		{ID: "aaa", OriginalPath: "/a", TrashedAt: mustParseTime("2026-01-01T00:00:00Z")},
		{ID: "bbb", OriginalPath: "/b", TrashedAt: mustParseTime("2026-01-02T00:00:00Z")},
	}

	if err := WriteAllEntries(indexPath, entries); err != nil {
		t.Fatalf("WriteAllEntries() returned error: %v", err)
	}

	got, err := ReadAllEntries(indexPath)
	if err != nil {
		t.Fatalf("ReadAllEntries() returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestAppendThenRewrite(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "trash.jsonl")

	e1 := &TrashEntry{ID: "aaa", OriginalPath: "/a", TrashedAt: mustParseTime("2026-01-01T00:00:00Z")}
	e2 := &TrashEntry{ID: "bbb", OriginalPath: "/b", TrashedAt: mustParseTime("2026-01-02T00:00:00Z")}

	AppendEntry(indexPath, e1)
	AppendEntry(indexPath, e2)

	all, _ := ReadAllEntries(indexPath)
	remaining := []*TrashEntry{all[0]}

	if err := WriteAllEntries(indexPath, remaining); err != nil {
		t.Fatalf("WriteAllEntries() returned error: %v", err)
	}

	got, _ := ReadAllEntries(indexPath)
	if len(got) != 1 || got[0].ID != "aaa" {
		t.Fatalf("expected 1 entry with ID aaa, got %d entries", len(got))
	}
}

func TestMalformedLine(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "trash.jsonl")

	if err := os.WriteFile(indexPath, []byte("valid line\n"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := ReadAllEntries(indexPath)
	if err != nil {
		t.Fatalf("ReadAllEntries() returned error: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("expected 0 entries (malformed line skipped), got %d", len(entries))
	}
}

func TestConcurrentAppend(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "trash.jsonl")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			e := &TrashEntry{
				ID:           strings.ToUpper(string(rune('A' + n))),
				OriginalPath: "/file",
				TrashedAt:    mustParseTime("2026-01-01T00:00:00Z"),
			}
			AppendEntry(indexPath, e)
		}(i)
	}
	wg.Wait()

	entries, err := ReadAllEntries(indexPath)
	if err != nil {
		t.Fatalf("ReadAllEntries() returned error: %v", err)
	}

	if len(entries) != 10 {
		t.Fatalf("expected 10 entries after concurrent append, got %d", len(entries))
	}
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
