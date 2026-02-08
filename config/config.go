package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	NotesPath   string   `json:"notes_path"`
	IgnoreDecks []string `json:"ignore_decks,omitempty"`
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

	_ = json.Unmarshal(data, &cfg)
	return cfg
}

// ResolvePath returns the provided path if non-empty, otherwise falls back to config.
func (c Config) ResolvePath(arg string) string {
	if arg != "" {
		return arg
	}
	return c.NotesPath
}

// Save writes the config back to the given path as indented JSON, creating the directory if needed.
func (c Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsDeckIgnored returns true if the deck name matches any entry in the ignore list.
// Matching is done by prefix, so "leetcode" ignores "leetcode", "leetcode.dp.tasks", etc.
func (c Config) IsDeckIgnored(deck string) bool {
	for _, pattern := range c.IgnoreDecks {
		if deck == pattern || strings.HasPrefix(deck, pattern+".") {
			return true
		}
	}
	return false
}
