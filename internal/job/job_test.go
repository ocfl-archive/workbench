package job

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
)

func TestGetBatches(t *testing.T) {
	tempDir := t.TempDir()

	if err := os.Mkdir(filepath.Join(tempDir, "batch_1"), 0755); err != nil {
		t.Fatalf("failed to create batch dir: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "batch_2"), 0755); err != nil {
		t.Fatalf("failed to create batch dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	batches, err := GetBatches(tempDir)
	if err != nil {
		t.Fatalf("GetBatches returned error: %v", err)
	}
	if len(batches) != 2 {
		t.Fatalf("expected 2 batches, got %d", len(batches))
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

	// Pre-create OCFL zip, PDF, and upload.json for job1
	if err := os.WriteFile(filepath.Join(ocflDir, "job1.zip"), []byte("ocfl-content"), 0644); err != nil {
		t.Fatalf("failed to write job1 ocfl: %v", err)
	}
	if err := os.WriteFile(filepath.Join(reportDir, "job1.pdf"), []byte("pdf-content"), 0644); err != nil {
		t.Fatalf("failed to write job1 report: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ocflDir, "job1.upload.json"), []byte(`{"date":"2026-10-08T12:00:00Z"}`), 0644); err != nil {
		t.Fatalf("failed to write job1 upload.json: %v", err)
	}

	logger := zerolog.Nop()
	jobs, err := GetSignatures(batchName, incomingDir, ocflDir, reportDir, logger)
	if err != nil {
		t.Fatalf("GetSignatures returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	j := jobs[0]
	if j.Batch != batchName {
		t.Errorf("job.Batch = %s, want %s", j.Batch, batchName)
	}
	if j.Signature != "SIG-001" {
		t.Errorf("job.Signature = %s, want SIG-001", j.Signature)
	}
	if j.Title != "Test Job 1" {
		t.Errorf("job.Title = %s, want Test Job 1", j.Title)
	}
	if j.OcflFile == "" {
		t.Errorf("expected job.OcflFile to be set")
	}
	if j.ReportFile == "" {
		t.Errorf("expected job.ReportFile to be set")
	}
	if j.UploadFile == "" {
		t.Errorf("expected job.UploadFile to be set")
	}
}
