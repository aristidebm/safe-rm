package engine

import (
	"os"
	"path/filepath"
	"strings"
)

func MatchesAny(path string, patterns []string) (bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	for _, pattern := range patterns {
		p := pattern

		if strings.HasPrefix(p, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return false, err
			}
			p = filepath.Join(home, p[2:])
		} else if p == "~" {
			home, err := os.UserHomeDir()
			if err != nil {
				return false, err
			}
			p = home
		}

		matched, err := match(abs, p)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

func match(path, pattern string) (bool, error) {
	if !strings.Contains(pattern, "**") {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}

		if !strings.Contains(pattern, string(filepath.Separator)) {
			base := filepath.Base(path)
			return filepath.Match(pattern, base)
		}

		return false, nil
	}

	parts := strings.SplitN(pattern, "**", 2)
	prefix := parts[0]
	suffix := parts[1]

	if prefix != "" && !strings.HasPrefix(path, prefix) {
		return false, nil
	}

	if suffix != "" && !strings.HasSuffix(path, suffix) {
		return false, nil
	}

	rest := path[len(prefix):]
	if suffix != "" {
		rest = rest[:len(rest)-len(suffix)]
	}

	if strings.Contains(rest, "/") {
		return true, nil
	}

	return true, nil
}
