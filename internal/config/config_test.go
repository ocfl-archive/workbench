package config

import (
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
