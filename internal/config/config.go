// Package config handles loading, parsing, and resolving TOML configuration files
// for the Workbench application suite. It provides dynamic batch path expansion
// and automatic selection of gocfl configuration profiles based on item file counts.
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
	Config      string `toml:"config"`
	LargeConfig string `toml:"largeconfig"`
	LargeLimit  int    `toml:"largelimit"`
}

// ONAConfig holds configuration specific to ona command-line executions.
type ONAConfig struct {
	Config string `toml:"config"`
}

// WBConfig holds the main configuration for the workbench environment.
type WBConfig struct {
	Batches    string      `toml:"batches"`
	Input      string      `toml:"input"`
	Ocfl       string      `toml:"ocfl"`
	Report     string      `toml:"report"`
	Error      string      `toml:"error"`
	Archived   string      `toml:"archived"`
	Gocfl      GOCFLConfig `toml:"gocfl"`
	Ona        ONAConfig   `toml:"ona"`
	ConfigFile string      `toml:"-"`
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

// CountFiles returns the total number of non-directory files found recursively within the specified folder.
// If the folder is empty string or does not exist, it returns 0.
func CountFiles(folder string) (int, error) {
	if folder == "" {
		return 0, nil
	}
	fi, err := os.Stat(folder)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	if !fi.IsDir() {
		return 0, fmt.Errorf("path '%s' is not a directory", folder)
	}
	count := 0
	err = filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetConfig returns the appropriate gocfl configuration file path based on the file count in dataFolder.
// If dataFolder contains more than LargeLimit files and LargeConfig is set, LargeConfig is returned;
// otherwise Config is returned.
func (c *GOCFLConfig) GetConfig(dataFolder string) string {
	if c == nil {
		return ""
	}
	if c.LargeLimit > 0 && c.LargeConfig != "" && dataFolder != "" {
		count, err := CountFiles(dataFolder)
		if err == nil && count > c.LargeLimit {
			return c.LargeConfig
		}
	}
	return c.Config
}

// GetGOCFLConfig returns the appropriate gocfl configuration file path based on the file count in dataFolder.
func (c *WBConfig) GetGOCFLConfig(dataFolder string) string {
	if c == nil {
		return ""
	}
	return c.Gocfl.GetConfig(dataFolder)
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

	conf.ConfigFile = configPath

	return &conf, nil
}
