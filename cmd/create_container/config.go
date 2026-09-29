package main

import (
	"fmt"
	"os"
	"path/filepath"
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

func (c *WBConfig) GetInputFolder(batch string) string {
	if c.Input != "" {
		if strings.Contains(c.Input, "{batch}") {
			return filepath.Clean(strings.ReplaceAll(c.Input, "{batch}", batch))
		}
		return filepath.Clean(c.Input)
	}
	if c.Batches != "" {
		return filepath.Join(c.Batches, batch, "incoming")
	}
	return ""
}

func (c *WBConfig) GetOcflFolder(batch string) string {
	if c.Ocfl != "" {
		if strings.Contains(c.Ocfl, "{batch}") {
			return filepath.Clean(strings.ReplaceAll(c.Ocfl, "{batch}", batch))
		}
		return filepath.Clean(c.Ocfl)
	}
	if c.Batches != "" {
		return filepath.Join(c.Batches, batch, "ocfl")
	}
	return ""
}

func (c *WBConfig) GetReportFolder(batch string) string {
	if c.Report != "" {
		if strings.Contains(c.Report, "{batch}") {
			return filepath.Clean(strings.ReplaceAll(c.Report, "{batch}", batch))
		}
		return filepath.Clean(c.Report)
	}
	if c.Batches != "" {
		return filepath.Join(c.Batches, batch, "report")
	}
	return ""
}

func (c *WBConfig) GetErrorFolder(batch string) string {
	if c.Error != "" {
		if strings.Contains(c.Error, "{batch}") {
			return filepath.Clean(strings.ReplaceAll(c.Error, "{batch}", batch))
		}
		return filepath.Clean(c.Error)
	}
	if c.Batches != "" {
		return filepath.Join(c.Batches, batch, "error")
	}
	return ""
}

func (c *WBConfig) GetArchivedFolder(batch string) string {
	if c.Archived != "" {
		if strings.Contains(c.Archived, "{batch}") {
			return filepath.Clean(strings.ReplaceAll(c.Archived, "{batch}", batch))
		}
		return filepath.Clean(c.Archived)
	}
	if c.Batches != "" {
		return filepath.Join(c.Batches, batch, "archived")
	}
	return ""
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
