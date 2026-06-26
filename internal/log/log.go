package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	mu     sync.Mutex
	logger *Logger
)

type Logger struct {
	w     io.WriteCloser
	level Level
}

func Init(debug bool) error {
	mu.Lock()
	defer mu.Unlock()

	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		stateDir = filepath.Join(home, ".local", "state")
	}

	logDir := filepath.Join(stateDir, "safe-rm")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(logDir, "safe-rm.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	lvl := LevelInfo
	if debug {
		lvl = LevelDebug
	}

	logger = &Logger{w: f, level: lvl}
	return nil
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logger != nil {
		logger.w.Close()
	}
}

func logf(lvl Level, format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()

	if logger == nil || lvl < logger.level {
		return
	}

	t := time.Now().Format("2006-01-02T15:04:05.000")
	label := [...]string{"DEBUG", "INFO ", "WARN ", "ERROR"}

	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(logger.w, "%s [%s] %s\n", t, label[lvl], msg)
}

func Debugf(format string, args ...any) { logf(LevelDebug, format, args...) }
func Infof(format string, args ...any)  { logf(LevelInfo, format, args...) }
func Warnf(format string, args ...any)  { logf(LevelWarn, format, args...) }
func Errorf(format string, args ...any) { logf(LevelError, format, args...) }
