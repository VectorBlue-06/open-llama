package config

import (
	"encoding/json"
	"errors"
	"os"
)

// Load reads the config from the default config path.
// If the file does not exist, it creates one with defaults and returns it.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom reads a config from the given path.
// If the file does not exist, it writes defaults to that path and returns them.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := Defaults()
			if saveErr := SaveTo(cfg, path); saveErr != nil {
				return cfg, nil // return defaults even if save fails
			}
			return cfg, nil
		}
		return nil, err
	}

	cfg := Defaults() // start from defaults so missing fields keep default values
	if err := json.Unmarshal(data, cfg); err != nil {
		// Invalid JSON — fall back to defaults
		return Defaults(), nil
	}
	return cfg, nil
}

// Save writes the config to the default config path.
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	return SaveTo(cfg, path)
}

// SaveTo writes the config to the given path.
func SaveTo(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
