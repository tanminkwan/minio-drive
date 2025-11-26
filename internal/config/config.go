package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type MinIOConfig struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Bucket    string `json:"bucket"`
	UseSSL    bool   `json:"use_ssl"`
}

type MountConfig struct {
	Type        string `json:"type"`         // "webdav" or "winfsp"
	Port        int    `json:"port"`         // WebDAV port (only for webdav)
	DriveLetter string `json:"drive_letter"`
	AutoStart   bool   `json:"auto_start"`
}

type Config struct {
	MinIO MinIOConfig `json:"minio"`
	Mount MountConfig `json:"mount"`
}

// IsWebDAV returns true if mount type is webdav
func (c *Config) IsWebDAV() bool {
	return c.Mount.Type == "" || c.Mount.Type == "webdav"
}

// IsWinFsp returns true if mount type is winfsp
func (c *Config) IsWinFsp() bool {
	return c.Mount.Type == "winfsp"
}

// GetConfigPath returns the path to config.json next to the executable
func GetConfigPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exePath), "config.json"), nil
}

// Load reads the configuration from config.json
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the configuration to config.json
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
