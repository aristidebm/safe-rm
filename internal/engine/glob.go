package engine

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func MatchesAny(path string, patterns []string) (bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	for _, pattern := range patterns {
		p := expandTilde(pattern)

		matched, err := doublestar.Match(p, abs)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}

		if !strings.Contains(p, string(filepath.Separator)) && !strings.Contains(p, "**") {
			base := filepath.Base(abs)
			matched, err := doublestar.Match(p, base)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
	}

	return false, nil
}

func expandTilde(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[2:])
		}
	} else if p == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	return p
}
