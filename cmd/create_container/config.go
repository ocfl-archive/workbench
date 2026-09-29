package main

import (
	"fmt"
	"os"
	"strings"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
)

var DefaultConfigPath = []string{"/etc/wb/workbench.toml", "/opt/wb/workbench.toml", "~/wb/workbench.toml"}

type GOCFLConfig struct {
	Config string `toml:"config"`
}

type WBConfig struct {
	Batches  string      `toml:"batches"`
	Input    string      `toml:"input"`
	Ocfl     string      `toml:"ocfl"`
	Report   string      `toml:"report"`
	Error    string      `toml:"error"`
	Archived string      `toml:"archived"`
	Gocfl    GOCFLConfig `toml:"gocfl"`
}

func LoadConfig(configPath string) (*WBConfig, error) {
	if configPath == "" {
		for _, path := range DefaultConfigPath {
			if strings.Contains(path, "~/") {
				path = os.ExpandEnv(path)
			}
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}
	if configPath == "" {
		return nil, fmt.Errorf("no config file found")
	}
	var config = &WBConfig{}
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return nil, errors.Wrapf(err, "failed to decode config file %s", configPath)
	}
	return config, nil
}
