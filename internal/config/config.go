package config

import (
	"encoding/json"
	"io"
	"os"
)

// repressent the structure of the gomon.json configuration file
type Config struct {
	Watch  []string          `json:"watch"`
	Ignore []string          `json:"ignore"`
	Build  BuildConfig       `json:"build"`
	Run    string            `json:"run"`
	Env    map[string]string `json:"env"`
}

type BuildConfig struct {
	Command   string `json:"command"`
	Directory string `json:"directory"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
