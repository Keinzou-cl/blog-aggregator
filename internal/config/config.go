package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configFile := filepath.Join(home, configFileName)
	return configFile, nil
}

func Read() (*Config, error) {
	configFile, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func write(cfg *Config) error {
	configFile, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return err
	}
	return nil
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	if err := write(c); err != nil {
		return err
	}
	return nil
}
