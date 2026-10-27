package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestWBConfigPathResolution(t *testing.T) {
	conf := &WBConfig{
		Batches:  filepath.Join("C:", "data", "batches"),
		Input:    filepath.Join("C:", "data", "batches", "{batch}", "incoming"),
		Ocfl:     filepath.Join("C:", "data", "batches", "{batch}", "ocfl"),
		Report:   filepath.Join("C:", "data", "batches", "{batch}", "report"),
		Error:    filepath.Join("C:", "data", "batches", "{batch}", "error"),
		Archived: filepath.Join("C:", "data", "batches", "{batch}", "archived"),
	}

	batch := "batch_01"
	if got := conf.GetInputFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "incoming") {
		t.Errorf("GetInputFolder() = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "incoming"))
	}
	if got := conf.GetOcflFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "ocfl") {
		t.Errorf("GetOcflFolder() = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "ocfl"))
	}
	if got := conf.GetReportFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "report") {
		t.Errorf("GetReportFolder() = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "report"))
	}
	if got := conf.GetErrorFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "error") {
		t.Errorf("GetErrorFolder() = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "error"))
	}
	if got := conf.GetArchivedFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "archived") {
		t.Errorf("GetArchivedFolder() = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "archived"))
	}

	// Test static paths without {batch} placeholder
	staticConf := &WBConfig{
		Batches:  filepath.Join("C:", "data", "batches"),
		Input:    filepath.Join("C:", "shared", "input"),
		Ocfl:     filepath.Join("C:", "shared", "ocfl"),
		Report:   filepath.Join("C:", "shared", "report"),
		Error:    filepath.Join("C:", "shared", "error"),
		Archived: filepath.Join("C:", "shared", "archived"),
	}
	if got := staticConf.GetInputFolder(batch); got != filepath.Join("C:", "shared", "input") {
		t.Errorf("GetInputFolder() static = %v, want %v", got, filepath.Join("C:", "shared", "input"))
	}
	if got := staticConf.GetOcflFolder(batch); got != filepath.Join("C:", "shared", "ocfl") {
		t.Errorf("GetOcflFolder() static = %v, want %v", got, filepath.Join("C:", "shared", "ocfl"))
	}
	if got := staticConf.GetReportFolder(batch); got != filepath.Join("C:", "shared", "report") {
		t.Errorf("GetReportFolder() static = %v, want %v", got, filepath.Join("C:", "shared", "report"))
	}
	if got := staticConf.GetErrorFolder(batch); got != filepath.Join("C:", "shared", "error") {
		t.Errorf("GetErrorFolder() static = %v, want %v", got, filepath.Join("C:", "shared", "error"))
	}
	if got := staticConf.GetArchivedFolder(batch); got != filepath.Join("C:", "shared", "archived") {
		t.Errorf("GetArchivedFolder() static = %v, want %v", got, filepath.Join("C:", "shared", "archived"))
	}

	// Test fallback when fields are empty
	emptyConf := &WBConfig{
		Batches: filepath.Join("C:", "data", "batches"),
	}
	if got := emptyConf.GetInputFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "incoming") {
		t.Errorf("GetInputFolder() fallback = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "incoming"))
	}
	if got := emptyConf.GetOcflFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "ocfl") {
		t.Errorf("GetOcflFolder() fallback = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "ocfl"))
	}
	if got := emptyConf.GetReportFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "report") {
		t.Errorf("GetReportFolder() fallback = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "report"))
	}
	if got := emptyConf.GetErrorFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "error") {
		t.Errorf("GetErrorFolder() fallback = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "error"))
	}
	if got := emptyConf.GetArchivedFolder(batch); got != filepath.Join("C:", "data", "batches", "batch_01", "archived") {
		t.Errorf("GetArchivedFolder() fallback = %v, want %v", got, filepath.Join("C:", "data", "batches", "batch_01", "archived"))
	}
}

func TestLoadConfigWithOna(t *testing.T) {
	tempDir := t.TempDir()
	tomlPath := filepath.Join(tempDir, "workbench.toml")
	tomlContent := `
batches = "C:/data/batches/"
input = "C:/data/batches/{batch}/incoming/"
ocfl = "C:/data/batches/{batch}/ocfl/"

[gocfl]
config = "C:/gocfl/config.toml"
largeconfig = "C:/gocfl/largeconfig.toml"
largelimit = 50

[ona]
config = "C:/ona/config.yml"
`
	if err := os.WriteFile(tomlPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test toml: %v", err)
	}

	conf, err := LoadConfig(tomlPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if conf.ConfigFile != tomlPath {
		t.Errorf("conf.ConfigFile = %v, want %v", conf.ConfigFile, tomlPath)
	}

	if conf.Gocfl.Config != "C:/gocfl/config.toml" {
		t.Errorf("conf.Gocfl.Config = %v, want C:/gocfl/config.toml", conf.Gocfl.Config)
	}
	if conf.Gocfl.LargeConfig != "C:/gocfl/largeconfig.toml" {
		t.Errorf("conf.Gocfl.LargeConfig = %v, want C:/gocfl/largeconfig.toml", conf.Gocfl.LargeConfig)
	}
	if conf.Gocfl.LargeLimit != 50 {
		t.Errorf("conf.Gocfl.LargeLimit = %v, want 50", conf.Gocfl.LargeLimit)
	}
	if conf.Ona.Config != "C:/ona/config.yml" {
		t.Errorf("conf.Ona.Config = %v, want C:/ona/config.yml", conf.Ona.Config)
	}
}

func TestGetGOCFLConfig(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data dir: %v", err)
	}

	conf := &WBConfig{
		Gocfl: GOCFLConfig{
			Config:      "C:/gocfl/normal.toml",
			LargeConfig: "C:/gocfl/large.toml",
			LargeLimit:  3,
		},
	}

	// 0 files: should return normal config
	if got := conf.GetGOCFLConfig(dataDir); got != "C:/gocfl/normal.toml" {
		t.Errorf("GetGOCFLConfig() with 0 files = %v, want %v", got, "C:/gocfl/normal.toml")
	}

	// 3 files (<= LargeLimit): should return normal config
	for i := 1; i <= 3; i++ {
		filePath := filepath.Join(dataDir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}
	if got := conf.GetGOCFLConfig(dataDir); got != "C:/gocfl/normal.toml" {
		t.Errorf("GetGOCFLConfig() with 3 files = %v, want %v", got, "C:/gocfl/normal.toml")
	}

	// 4 files (> LargeLimit): should return large config
	subDir := filepath.Join(dataDir, "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file4.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	if got := conf.GetGOCFLConfig(dataDir); got != "C:/gocfl/large.toml" {
		t.Errorf("GetGOCFLConfig() with 4 files = %v, want %v", got, "C:/gocfl/large.toml")
	}

	// Non-existent data folder: should fallback to normal config
	if got := conf.GetGOCFLConfig(filepath.Join(tempDir, "nonexistent")); got != "C:/gocfl/normal.toml" {
		t.Errorf("GetGOCFLConfig() nonexistent folder = %v, want %v", got, "C:/gocfl/normal.toml")
	}

	// Empty data folder path: should return normal config
	if got := conf.GetGOCFLConfig(""); got != "C:/gocfl/normal.toml" {
		t.Errorf("GetGOCFLConfig(\"\") = %v, want %v", got, "C:/gocfl/normal.toml")
	}

	// Nil config: should return empty string
	var nilConf *WBConfig
	if got := nilConf.GetGOCFLConfig(dataDir); got != "" {
		t.Errorf("nilConf.GetGOCFLConfig() = %v, want empty string", got)
	}

	// LargeConfig is empty: should fallback to Config even if count > LargeLimit
	confNoLargeConfig := &WBConfig{
		Gocfl: GOCFLConfig{
			Config:     "C:/gocfl/normal.toml",
			LargeLimit: 2,
		},
	}
	if got := confNoLargeConfig.GetGOCFLConfig(dataDir); got != "C:/gocfl/normal.toml" {
		t.Errorf("confNoLargeConfig.GetGOCFLConfig() = %v, want %v", got, "C:/gocfl/normal.toml")
	}

	// LargeLimit is 0: should return normal Config
	confZeroLimit := &WBConfig{
		Gocfl: GOCFLConfig{
			Config:      "C:/gocfl/normal.toml",
			LargeConfig: "C:/gocfl/large.toml",
			LargeLimit:  0,
		},
	}
	if got := confZeroLimit.GetGOCFLConfig(dataDir); got != "C:/gocfl/normal.toml" {
		t.Errorf("confZeroLimit.GetGOCFLConfig() = %v, want %v", got, "C:/gocfl/normal.toml")
	}
}
