package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type ThemeConfig struct {
	TitleFG      string `toml:"title_fg"`
	DangerFG     string `toml:"danger_fg"`
	WarningFG    string `toml:"warning_fg"`
	MutedFG      string `toml:"muted_fg"`
	SelectedFG   string `toml:"selected_fg"`
	UnselectedFG string `toml:"unselected_fg"`
	PermanentFG  string `toml:"permanent_fg"`
	PermanentBG  string `toml:"permanent_bg"`
	TrashFG      string `toml:"trash_fg"`
	TrashBG      string `toml:"trash_bg"`
	TrashPathFG  string `toml:"trash_path_fg"`
	BorderColor  string `toml:"border_color"`
	KeyHintFG    string `toml:"keyhint_fg"`
}

type Config struct {
	TrashDir   *string      `toml:"trash_dir"`
	BypassList []string     `toml:"bypass_list"`
	DangerList []string     `toml:"danger_list"`
	MaxAge     string       `toml:"max_age"`
	Theme      *ThemeConfig `toml:"theme"`
}

func configFilePath() (string, error) {
	if p := os.Getenv("SAFE_RM_CONFIG"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configHome = filepath.Join(home, ".config")
	}

	p := filepath.Join(configHome, "safe-rm", "config.toml")
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}

	return "", nil
}

func Load() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	if path == "" {
		return &Config{}, nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("config: %s: %w", path, err)
	}

	if cfg.TrashDir != nil {
		expanded := expandPath(*cfg.TrashDir)
		cfg.TrashDir = &expanded
	}

	return &cfg, nil
}

func (c *Config) ResolvedTrashDir() (string, error) {
	if c.TrashDir != nil {
		return *c.TrashDir, nil
	}

	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	return filepath.Join(dataHome, "Trash"), nil
}

func (c *Config) DefaultMaxAge() string {
	if c.MaxAge != "" {
		return c.MaxAge
	}
	return ""
}

func IndexDir() (string, error) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	return filepath.Join(dataHome, "safe-rm"), nil
}

func expandPath(s string) string {
	if strings.HasPrefix(s, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if len(s) == 1 || s[1] == '/' {
				s = filepath.Join(home, s[1:])
			}
		}
	}

	return os.ExpandEnv(s)
}
