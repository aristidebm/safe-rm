package engine

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"example.com/safe-rm/internal/log"
	"golang.org/x/sys/unix"
)

const lockTimeout = 2 * time.Second

type TrashEntry struct {
	ID           string    `json:"id"`
	OriginalPath string    `json:"original_path"`
	TrashedAt    time.Time `json:"trashed_at"`
	Size         int64     `json:"size"`
	IsDir        bool      `json:"is_dir"`
	Checksum     string    `json:"checksum"`
	Permanent    bool      `json:"permanent"`
	FreeDesktop  bool      `json:"free_desktop"`
}

func NewEntry(originalPath string, isDir bool) (*TrashEntry, error) {
	id, err := randomID(6)
	if err != nil {
		return nil, err
	}

	e := &TrashEntry{
		ID:           id,
		OriginalPath: originalPath,
		TrashedAt:    time.Now().UTC(),
		IsDir:        isDir,
	}

	if !isDir {
		cksum, err := fileChecksum(originalPath)
		if err != nil {
			return nil, err
		}
		e.Checksum = cksum

		fi, err := os.Stat(originalPath)
		if err != nil {
			return nil, err
		}
		e.Size = fi.Size()
	}

	return e, nil
}

func AppendEntry(indexPath string, e *TrashEntry) error {
	f, err := openWithLock(indexPath)
	if err != nil {
		return err
	}
	defer f.Close()
	defer unix.Flock(int(f.Fd()), unix.LOCK_UN)

	line, err := json.Marshal(e)
	if err != nil {
		return err
	}

	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}

	return nil
}

func ReadAllEntries(indexPath string) ([]*TrashEntry, error) {
	f, err := os.OpenFile(indexPath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []*TrashEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e TrashEntry
		if err := json.Unmarshal(line, &e); err != nil {
			log.Warnf("skipping malformed JSONL line: %v", err)
			continue
		}
		entries = append(entries, &e)
	}

	return entries, scanner.Err()
}

func WriteAllEntries(indexPath string, entries []*TrashEntry) error {
	f, err := openWithLock(indexPath)
	if err != nil {
		return err
	}
	defer f.Close()
	defer unix.Flock(int(f.Fd()), unix.LOCK_UN)

	tmpPath := indexPath + ".tmp"
	if err := writeJSONL(tmpPath, entries); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, indexPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

func randomID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func openWithLock(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(lockTimeout)
	for {
		if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err == nil {
			return f, nil
		} else if err != unix.EWOULDBLOCK {
			f.Close()
			return nil, err
		}

		if time.Now().After(deadline) {
			f.Close()
			return nil, fmt.Errorf("trash index is locked by another process")
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func writeJSONL(path string, entries []*TrashEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, e := range entries {
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
	}

	return nil
}
