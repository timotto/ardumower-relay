package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

func Get(osArgs []string) (*Configuration, error) {
	if len(osArgs) > 1 {
		return ReadConfig(Filename(osArgs))
	} else {
		return DefaultConfig()
	}
}

func ReadConfig(filename string) (*Configuration, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file %v: %w", filename, err)
	}

	cfg := &Configuration{}
	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode configuration file %v: %w", filename, err)
	}

	return cfg, cfg.Validate()
}

func DefaultConfig() (*Configuration, error) {
	cfg := &Configuration{}
	cfg.Server.Http.Address = ":8080"

	return cfg, cfg.Validate()
}

func Filename(osArgs []string) string {
	if len(osArgs) == 2 {
		return osArgs[1]
	}

	return "config.yml"
}
