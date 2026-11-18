package job

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ocfl-archive/workbench/internal/config"
)

func TestCheckIngestResult_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		content    string
		hasError   bool
		wantErrStr bool
	}{
		{
			name:       "Clean result with empty string error",
			content:    `{"status": "ok", "error": ""}`,
			hasError:   false,
			wantErrStr: false,
		},
		{
			name:       "Clean result with null error",
			content:    `{"status": "ok", "error": null}`,
			hasError:   false,
			wantErrStr: false,
		},
		{
			name:       "Clean result without error field",
			content:    `{"status": "ok", "uploaded": true}`,
			hasError:   false,
			wantErrStr: false,
		},
		{
			name:       "Error with string message",
			content:    `{"status": "failed", "error": "checksum mismatch"}`,
			hasError:   true,
			wantErrStr: true,
		},
		{
			name:       "Error with boolean true",
			content:    `{"status": "failed", "error": true}`,
			hasError:   true,
			wantErrStr: true,
		},
		{
			name:       "Clean with boolean false",
			content:    `{"status": "ok", "error": false}`,
			hasError:   false,
			wantErrStr: false,
		},
		{
			name:       "Error with structured object",
			content:    `{"status": "failed", "error": {"code": 500, "detail": "server error"}}`,
			hasError:   true,
			wantErrStr: true,
		},
		{
			name:       "Error with list of issues",
			content:    `{"status": "failed", "error": ["invalid schema", "missing field"]}`,
			hasError:   true,
			wantErrStr: true,
		},
		{
			name:       "Malformed JSON",
			content:    `{invalid json`,
			hasError:   true,
			wantErrStr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, "result.json")
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			hasError, errorMsg, _ := CheckIngestResult(filePath)
			if hasError != tt.hasError {
				t.Errorf("hasError = %v, want %v", hasError, tt.hasError)
			}
			if (errorMsg != "") != tt.wantErrStr {
				t.Errorf("errorMsg = %q (non-empty = %v), wantErrStr = %v", errorMsg, errorMsg != "", tt.wantErrStr)
			}
		})
	}

	t.Run("Missing file", func(t *testing.T) {
		missingPath := filepath.Join(tempDir, "nonexistent.json")
		hasError, errorMsg, err := CheckIngestResult(missingPath)
		if !hasError {
			t.Errorf("expected hasError = true for missing file")
		}
		if errorMsg == "" {
			t.Errorf("expected non-empty errorMsg for missing file")
		}
		if err == nil {
			t.Errorf("expected non-nil err for missing file")
		}
	})
}

func TestCreateUploadRecord_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	batchDir := filepath.Join(tempDir, "batch_u")
	if err := os.MkdirAll(batchDir, 0755); err != nil {
		t.Fatalf("failed to create batch dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	j := Job{
		BaseName:  "item_upload",
		Batch:     "batch_u",
		Signature: "SIG-U",
	}

	if err := CreateUploadRecord(conf, j); err != nil {
		t.Fatalf("CreateUploadRecord failed: %v", err)
	}

	uploadPath := j.ExpectedUploadPath(conf)
	data, err := os.ReadFile(uploadPath)
	if err != nil {
		t.Fatalf("failed to read upload file: %v", err)
	}

	var uploadInfo UploadInfo
	if err := json.Unmarshal(data, &uploadInfo); err != nil {
		t.Fatalf("failed to unmarshal upload info: %v", err)
	}

	parsedTime, err := time.Parse(time.RFC3339, uploadInfo.Date)
	if err != nil {
		t.Fatalf("failed to parse date: %v", err)
	}
	if time.Since(parsedTime) > 5*time.Second {
		t.Errorf("upload timestamp is too old: %v", parsedTime)
	}
}

func TestHandleJobError_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "input"),
		Error:    filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:   filepath.Join(tempDir, "{batch}", "report"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	batch := "batch_1"
	inputDir := conf.GetInputFolder(batch)
	jobDir := filepath.Join(inputDir, "item_err")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}

	ocflDir := conf.GetOcflFolder(batch)
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}
	zipPath := filepath.Join(ocflDir, "item_err.zip")
	if err := os.WriteFile(zipPath, []byte("partial zip"), 0644); err != nil {
		t.Fatalf("failed to write partial zip: %v", err)
	}
	checksumPath := filepath.Join(ocflDir, "item_err.zip.sha512")
	if err := os.WriteFile(checksumPath, []byte("partial checksum"), 0644); err != nil {
		t.Fatalf("failed to write partial checksum: %v", err)
	}

	j := Job{
		BaseName:       "item_err",
		Batch:          batch,
		IncomingFolder: jobDir,
		Signature:      "SIG-ERR",
	}

	if err := HandleJobError(conf, j, "gocfl create", "something went wrong"); err != nil {
		t.Fatalf("HandleJobError failed: %v", err)
	}

	// Verify job directory moved to error
	errDir := conf.GetErrorFolder(batch)
	dstJobDir := filepath.Join(errDir, "item_err")
	if fi, err := os.Stat(dstJobDir); err != nil || !fi.IsDir() {
		t.Errorf("job dir was not moved to error folder: %v", err)
	}

	// Verify log file was written
	logPath := filepath.Join(errDir, "item_err.log")
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read error log: %v", err)
	}
	if string(logData) != "gocfl create\n\nsomething went wrong" {
		t.Errorf("unexpected log content: %s", string(logData))
	}

	// Verify partial zip was removed
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Errorf("expected partial zip to be removed, but stat returned: %v", err)
	}

	// Verify partial checksum was removed
	if _, err := os.Stat(checksumPath); !os.IsNotExist(err) {
		t.Errorf("expected partial checksum to be removed, but stat returned: %v", err)
	}
}

func TestArchiveJob_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "input"),
		Error:    filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:   filepath.Join(tempDir, "{batch}", "report"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	batch := "batch_1"
	inputDir := conf.GetInputFolder(batch)
	jobDir := filepath.Join(inputDir, "item_arch")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}

	ocflDir := conf.GetOcflFolder(batch)
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}
	zipPath := filepath.Join(ocflDir, "item_arch.zip")
	if err := os.WriteFile(zipPath, []byte("ocfl zip content"), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}
	checksumPath := filepath.Join(ocflDir, "item_arch.zip.sha512")
	if err := os.WriteFile(checksumPath, []byte("checksum content"), 0644); err != nil {
		t.Fatalf("failed to write checksum: %v", err)
	}
	statusJsonPath := filepath.Join(ocflDir, "item_arch_ona.json")
	if err := os.WriteFile(statusJsonPath, []byte(`{"id":"ingest-123"}`), 0644); err != nil {
		t.Fatalf("failed to write status json: %v", err)
	}

	j := Job{
		BaseName:       "item_arch",
		Batch:          batch,
		IncomingFolder: jobDir,
		Signature:      "SIG-ARCH",
	}

	if err := ArchiveJob(conf, j, statusJsonPath); err != nil {
		t.Fatalf("ArchiveJob failed: %v", err)
	}

	archDir := conf.GetArchivedFolder(batch)
	dstJobDir := filepath.Join(archDir, "item_arch")
	if fi, err := os.Stat(dstJobDir); err != nil || !fi.IsDir() {
		t.Errorf("job dir was not moved to archived folder: %v", err)
	}
	dstZip := filepath.Join(archDir, "item_arch.zip")
	if fi, err := os.Stat(dstZip); err != nil || fi.IsDir() {
		t.Errorf("zip was not moved to archived folder: %v", err)
	}
	dstChecksum := filepath.Join(archDir, "item_arch.zip.sha512")
	if fi, err := os.Stat(dstChecksum); err != nil || fi.IsDir() {
		t.Errorf("checksum was not moved to archived folder: %v", err)
	}
	dstStatus := filepath.Join(archDir, "item_arch_ona.json")
	if fi, err := os.Stat(dstStatus); err != nil || fi.IsDir() {
		t.Errorf("status JSON was not moved to archived folder: %v", err)
	}
}

func TestResetJob_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "input"),
		Error:    filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:   filepath.Join(tempDir, "{batch}", "report"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	batch := "batch_reset"
	errDir := conf.GetErrorFolder(batch)
	jobDir := filepath.Join(errDir, "item_reset")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir in error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(jobDir, "info.json"), []byte(`{"signature":"SIG-RESET","title":"Reset Item"}`), 0644); err != nil {
		t.Fatalf("failed to write info.json: %v", err)
	}

	// External info json
	extInfo := filepath.Join(errDir, "item_reset.json")
	if err := os.WriteFile(extInfo, []byte(`{"signature":"SIG-RESET"}`), 0644); err != nil {
		t.Fatalf("failed to write ext info: %v", err)
	}

	// Error log
	errLog := filepath.Join(errDir, "item_reset.log")
	if err := os.WriteFile(errLog, []byte("error output log"), 0644); err != nil {
		t.Fatalf("failed to write error log: %v", err)
	}

	// Ingest status json in error
	errStatus := filepath.Join(errDir, "item_reset_ona.json")
	if err := os.WriteFile(errStatus, []byte(`{"error":"failed"}`), 0644); err != nil {
		t.Fatalf("failed to write error status: %v", err)
	}

	// Leftover OCFL files
	ocflDir := conf.GetOcflFolder(batch)
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}
	zipPath := filepath.Join(ocflDir, "item_reset.zip")
	if err := os.WriteFile(zipPath, []byte("leftover zip"), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}
	csPath := filepath.Join(ocflDir, "item_reset.zip.sha512")
	if err := os.WriteFile(csPath, []byte("leftover cs"), 0644); err != nil {
		t.Fatalf("failed to write cs: %v", err)
	}
	upPath := filepath.Join(ocflDir, "item_reset.upload.json")
	if err := os.WriteFile(upPath, []byte("leftover upload"), 0644); err != nil {
		t.Fatalf("failed to write upload: %v", err)
	}

	// Leftover report file
	repDir := conf.GetReportFolder(batch)
	if err := os.MkdirAll(repDir, 0755); err != nil {
		t.Fatalf("failed to create report dir: %v", err)
	}
	repPath := filepath.Join(repDir, "item_reset.pdf")
	if err := os.WriteFile(repPath, []byte("leftover pdf"), 0644); err != nil {
		t.Fatalf("failed to write pdf: %v", err)
	}

	j := Job{
		BaseName:       "item_reset",
		Batch:          batch,
		Status:         StatusError,
		IncomingFolder: jobDir,
		ErrorFile:      errLog,
		OcflFile:       zipPath,
		ChecksumFile:   csPath,
		ReportFile:     repPath,
		UploadFile:     upPath,
		Signature:      "SIG-RESET",
		Title:          "Reset Item",
	}

	if err := ResetJob(conf, j); err != nil {
		t.Fatalf("ResetJob failed: %v", err)
	}

	// Verify input folder has the job dir
	inputDir := conf.GetInputFolder(batch)
	dstJobDir := filepath.Join(inputDir, "item_reset")
	if fi, err := os.Stat(dstJobDir); err != nil || !fi.IsDir() {
		t.Errorf("job dir was not moved to input folder: %v", err)
	}

	// Verify external info json moved to input
	dstExtInfo := filepath.Join(inputDir, "item_reset.json")
	if fi, err := os.Stat(dstExtInfo); err != nil || fi.IsDir() {
		t.Errorf("external info was not moved to input folder: %v", err)
	}

	// Verify error folder artifacts cleaned up
	if _, err := os.Stat(jobDir); !os.IsNotExist(err) {
		t.Errorf("expected error job dir to be removed, got: %v", err)
	}
	if _, err := os.Stat(errLog); !os.IsNotExist(err) {
		t.Errorf("expected error log to be removed, got: %v", err)
	}
	if _, err := os.Stat(errStatus); !os.IsNotExist(err) {
		t.Errorf("expected error status json to be removed, got: %v", err)
	}
	if _, err := os.Stat(extInfo); !os.IsNotExist(err) {
		t.Errorf("expected error ext info to be removed, got: %v", err)
	}

	// Verify ocfl and report artifacts cleaned up
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Errorf("expected ocfl zip to be removed, got: %v", err)
	}
	if _, err := os.Stat(csPath); !os.IsNotExist(err) {
		t.Errorf("expected checksum to be removed, got: %v", err)
	}
	if _, err := os.Stat(upPath); !os.IsNotExist(err) {
		t.Errorf("expected upload marker to be removed, got: %v", err)
	}
	if _, err := os.Stat(repPath); !os.IsNotExist(err) {
		t.Errorf("expected report to be removed, got: %v", err)
	}
}
