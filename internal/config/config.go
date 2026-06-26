package config

type Config struct {
	TrashDir   *string  `toml:"trash_dir"`
	BypassList []string `toml:"bypass_list"`
	DangerList []string `toml:"danger_list"`
}

func Load() (*Config, error) {
	return &Config{}, nil
}

func (c *Config) ResolvedTrashDir() (string, error) {
	return "", nil
}

func IndexDir() (string, error) {
	return "", nil
}
