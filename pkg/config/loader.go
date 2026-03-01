package config

import (
	"conduit/pkg/api"
	"encoding/json"
	"os"
)

func LoadFromFile(path string) (*api.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg api.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
