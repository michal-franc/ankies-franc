package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	NotesPath string `json:"notes_path"`
}

func DefaultConfigPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "ankies-franc", "config.json")
}

func Load() Config {
	var cfg Config

	data, err := os.ReadFile(DefaultConfigPath())
	if err != nil {
		return cfg
	}

	json.Unmarshal(data, &cfg)
	return cfg
}

// ResolvePath returns the provided path if non-empty, otherwise falls back to config.
func (c Config) ResolvePath(arg string) string {
	if arg != "" {
		return arg
	}
	return c.NotesPath
}
