package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesLogFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	if err := Init(false); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer Close()

	logPath := filepath.Join(dir, "safe-rm", "safe-rm.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("log file was not created")
	}
}

func TestInfoWritesToLog(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	if err := Init(false); err != nil {
		t.Fatal(err)
	}
	defer Close()

	Infof("test message %d", 42)

	logPath := filepath.Join(dir, "safe-rm", "safe-rm.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "test message 42") {
		t.Fatalf("expected log to contain 'test message 42', got %q", content)
	}
	if !strings.Contains(content, "INFO ") {
		t.Fatalf("expected log to contain INFO level, got %q", content)
	}
}

func TestDebugNotLoggedWhenNotDebug(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	if err := Init(false); err != nil {
		t.Fatal(err)
	}
	defer Close()

	Debugf("debug message")

	logPath := filepath.Join(dir, "safe-rm", "safe-rm.log")
	data, _ := os.ReadFile(logPath)

	if len(data) > 0 {
		t.Fatalf("expected no log output in non-debug mode, got %q", string(data))
	}
}

func TestDebugLoggedWhenDebug(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	if err := Init(true); err != nil {
		t.Fatal(err)
	}
	defer Close()

	Debugf("debug message")

	logPath := filepath.Join(dir, "safe-rm", "safe-rm.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "debug message") {
		t.Fatalf("expected log to contain 'debug message', got %q", content)
	}
	if !strings.Contains(content, "DEBUG") {
		t.Fatalf("expected log to contain DEBUG level, got %q", content)
	}
}

func TestWarnAndErrorLevels(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	if err := Init(false); err != nil {
		t.Fatal(err)
	}
	defer Close()

	Warnf("warning message")
	Errorf("error message")

	logPath := filepath.Join(dir, "safe-rm", "safe-rm.log")
	data, _ := os.ReadFile(logPath)
	content := string(data)

	if !strings.Contains(content, "WARN ") {
		t.Fatalf("expected WARN in log, got %q", content)
	}
	if !strings.Contains(content, "ERROR") {
		t.Fatalf("expected ERROR in log, got %q", content)
	}
}
