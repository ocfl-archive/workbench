package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
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

func TestEnsureTargetDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Case 1: Directory already exists
	existingDir := filepath.Join(tempDir, "existing")
	if err := os.Mkdir(existingDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := EnsureTargetDirectory(existingDir); err != nil {
		t.Errorf("EnsureTargetDirectory(existingDir) returned unexpected error: %v", err)
	}

	// Case 2: Leaf directory missing, parent exists -> should succeed and create
	leafDir := filepath.Join(tempDir, "new_leaf")
	if err := EnsureTargetDirectory(leafDir); err != nil {
		t.Errorf("EnsureTargetDirectory(leafDir) returned unexpected error: %v", err)
	}
	if fi, err := os.Stat(leafDir); err != nil || !fi.IsDir() {
		t.Errorf("EnsureTargetDirectory(leafDir) did not create directory")
	}

	// Case 3: Parent and leaf missing -> should fail
	deepDir := filepath.Join(tempDir, "non_existent_parent", "target")
	if err := EnsureTargetDirectory(deepDir); err == nil {
		t.Errorf("EnsureTargetDirectory(deepDir) expected error for missing parent, got nil")
	}

	// Case 4: Target exists but is a file
	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := EnsureTargetDirectory(filePath); err == nil {
		t.Errorf("EnsureTargetDirectory(filePath) expected error for regular file, got nil")
	}
}

func TestGetSignaturesWithBatch(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_test"
	incomingDir := filepath.Join(tempDir, batchName, "incoming")
	ocflDir := filepath.Join(tempDir, batchName, "ocfl")
	reportDir := filepath.Join(tempDir, batchName, "report")

	if err := os.MkdirAll(incomingDir, 0755); err != nil {
		t.Fatalf("failed to create incoming dir: %v", err)
	}
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		t.Fatalf("failed to create report dir: %v", err)
	}

	// Setup job 1
	job1Dir := filepath.Join(incomingDir, "job1")
	if err := os.MkdirAll(filepath.Join(job1Dir, "data"), 0755); err != nil {
		t.Fatalf("failed to create job1 data: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(job1Dir, "metadata"), 0755); err != nil {
		t.Fatalf("failed to create job1 metadata: %v", err)
	}
	infoContent := `{"signature":"SIG-001","title":"Test Job 1"}`
	if err := os.WriteFile(filepath.Join(job1Dir, "info.json"), []byte(infoContent), 0644); err != nil {
		t.Fatalf("failed to write job1 info: %v", err)
	}

	// Pre-create OCFL zip and PDF for job1
	if err := os.WriteFile(filepath.Join(ocflDir, "job1.zip"), []byte("ocfl-content"), 0644); err != nil {
		t.Fatalf("failed to write job1 ocfl: %v", err)
	}
	if err := os.WriteFile(filepath.Join(reportDir, "job1.pdf"), []byte("pdf-content"), 0644); err != nil {
		t.Fatalf("failed to write job1 report: %v", err)
	}

	logger := zerolog.Nop()
	jobs, err := getSignatures(batchName, incomingDir, ocflDir, reportDir, logger)
	if err != nil {
		t.Fatalf("getSignatures returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	j := jobs[0]
	if j.batch != batchName {
		t.Errorf("job.batch = %s, want %s", j.batch, batchName)
	}
	if j.signature != "SIG-001" {
		t.Errorf("job.signature = %s, want SIG-001", j.signature)
	}
	if j.title != "Test Job 1" {
		t.Errorf("job.title = %s, want Test Job 1", j.title)
	}
	if j.ocflFile == "" {
		t.Errorf("expected job.ocflFile to be set")
	}
	if j.reportFile == "" {
		t.Errorf("expected job.reportFile to be set")
	}
}
