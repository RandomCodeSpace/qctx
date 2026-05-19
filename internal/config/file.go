package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

type fileConfig struct {
	SonarURL    string `yaml:"sonar_url"`
	SonarToken  string `yaml:"sonar_token"`
	GitLabURL   string `yaml:"gitlab_url"`
	GitLabToken string `yaml:"gitlab_token"`
	CACertPath  string `yaml:"ca_cert"`
	Insecure    bool   `yaml:"insecure"`
}

func loadFile(path string) (fileConfig, error) {
	if path == "" {
		return fileConfig{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fileConfig{}, nil
		}
		return fileConfig{}, fmt.Errorf("read config %q: %w", path, err)
	}
	var fc fileConfig
	if err := yaml.Unmarshal(b, &fc); err != nil {
		return fileConfig{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	return fc, nil
}
