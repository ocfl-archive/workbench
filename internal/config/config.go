package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
)

// GOCFLConfig holds configuration specific to gocfl command-line executions.
type GOCFLConfig struct {
	Config string `toml:"config"`
}

// ONAConfig holds configuration specific to ona command-line executions.
type ONAConfig struct {
	Config string `toml:"config"`
}

// WBConfig holds the main configuration for the workbench environment.
type WBConfig struct {
	Batches  string      `toml:"batches"`
	Input    string      `toml:"input"`
	Ocfl     string      `toml:"ocfl"`
	Report   string      `toml:"report"`
	Error    string      `toml:"error"`
	Archived string      `toml:"archived"`
	Gocfl    GOCFLConfig `toml:"gocfl"`
	Ona      ONAConfig   `toml:"ona"`
}

func (c *WBConfig) resolvePath(pattern, batch, fallbackSuffix string) string {
	if pattern != "" {
		if strings.Contains(pattern, "{batch}") {
			return filepath.Clean(strings.ReplaceAll(pattern, "{batch}", batch))
		}
		return filepath.Clean(pattern)
	}
	return filepath.Join(c.Batches, batch, fallbackSuffix)
}

// GetInputFolder returns the incoming/input directory path for the given batch.
func (c *WBConfig) GetInputFolder(batch string) string {
	return c.resolvePath(c.Input, batch, "incoming")
}

// GetOcflFolder returns the OCFL container destination directory for the given batch.
func (c *WBConfig) GetOcflFolder(batch string) string {
	return c.resolvePath(c.Ocfl, batch, "ocfl")
}

// GetReportFolder returns the report output directory for the given batch.
func (c *WBConfig) GetReportFolder(batch string) string {
	return c.resolvePath(c.Report, batch, "report")
}

// GetErrorFolder returns the error log directory for the given batch.
func (c *WBConfig) GetErrorFolder(batch string) string {
	return c.resolvePath(c.Error, batch, "error")
}

// GetArchivedFolder returns the archive destination directory for the given batch.
func (c *WBConfig) GetArchivedFolder(batch string) string {
	return c.resolvePath(c.Archived, batch, "archived")
}

// LoadConfig loads and decodes the workbench configuration from the specified file path,
// or falls back to default system search locations if configPath is empty.
func LoadConfig(configPath string) (*WBConfig, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(homeDir, "wb", "workbench.toml")
			if _, err := os.Stat(configPath); err != nil {
				configPath = ""
			}
		}
		if configPath == "" {
			if _, err := os.Stat("/opt/wb/workbench.toml"); err == nil {
				configPath = "/opt/wb/workbench.toml"
			} else if _, err := os.Stat("/etc/wb/workbench.toml"); err == nil {
				configPath = "/etc/wb/workbench.toml"
			}
		}
	}

	if configPath == "" {
		return nil, fmt.Errorf("no config file found")
	}

	var conf WBConfig
	if _, err := toml.DecodeFile(configPath, &conf); err != nil {
		return nil, errors.Wrapf(err, "error decoding config file '%s'", configPath)
	}

	return &conf, nil
}
