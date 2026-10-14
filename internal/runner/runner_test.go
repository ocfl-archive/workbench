package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/ocfl-archive/workbench/internal/job"
)

func TestIngestOCFLCmd(t *testing.T) {
	tempDir := t.TempDir()
	ocflDir := filepath.Join(tempDir, "batch_1", "ocfl")
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
		Ona: config.ONAConfig{
			Config: "C:/path/to/ona.yml",
		},
	}

	j := job.Job{
		BaseName:  "item_123",
		Batch:     "batch_1",
		Signature: "SIG-123",
	}

	cmd, resultJsonPath, err := IngestOCFLCmd(conf, j)
	if err != nil {
		t.Fatalf("IngestOCFLCmd returned error: %v", err)
	}

	expectedZipPath := filepath.Join(ocflDir, "item_123.zip")
	expectedJsonPath := filepath.Join(ocflDir, "item_123_ona.json")

	if resultJsonPath != expectedJsonPath {
		t.Errorf("resultJsonPath = %q, want %q", resultJsonPath, expectedJsonPath)
	}

	expectedArgs := []string{"ona", "ingest", "--status", expectedJsonPath, "-p", expectedZipPath, "-c", "C:/path/to/ona.yml"}

	if len(cmd.Args) != len(expectedArgs) {
		t.Fatalf("cmd.Args len = %d, want %d: %v", len(cmd.Args), len(expectedArgs), cmd.Args)
	}
	for i := range expectedArgs {
		if cmd.Args[i] != expectedArgs[i] {
			t.Errorf("cmd.Args[%d] = %s, want %s", i, cmd.Args[i], expectedArgs[i])
		}
	}
}

func TestIngestOCFLCmdWithoutConfig(t *testing.T) {
	tempDir := t.TempDir()
	ocflDir := filepath.Join(tempDir, "batch_2", "ocfl")
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	j := job.Job{
		BaseName:  "item_456",
		Batch:     "batch_2",
		Signature: "SIG-456",
	}

	cmd, resultJsonPath, err := IngestOCFLCmd(conf, j)
	if err != nil {
		t.Fatalf("IngestOCFLCmd returned error: %v", err)
	}

	expectedZipPath := filepath.Join(ocflDir, "item_456.zip")
	expectedJsonPath := filepath.Join(ocflDir, "item_456_ona.json")

	if resultJsonPath != expectedJsonPath {
		t.Errorf("resultJsonPath = %q, want %q", resultJsonPath, expectedJsonPath)
	}

	expectedArgs := []string{"ona", "ingest", "--status", expectedJsonPath, "-p", expectedZipPath}

	if len(cmd.Args) != len(expectedArgs) {
		t.Fatalf("cmd.Args len = %d, want %d: %v", len(cmd.Args), len(expectedArgs), cmd.Args)
	}
	for i := range expectedArgs {
		if cmd.Args[i] != expectedArgs[i] {
			t.Errorf("cmd.Args[%d] = %s, want %s", i, cmd.Args[i], expectedArgs[i])
		}
	}
}

func TestIngestOCFLCmdNilConfig(t *testing.T) {
	j := job.Job{
		BaseName:  "item_nil",
		Batch:     "batch_nil",
		Signature: "SIG-NIL",
	}

	_, _, err := IngestOCFLCmd(nil, j)
	if err == nil {
		t.Errorf("expected error when conf is nil, got nil")
	}
}

func TestCheckIngestResult(t *testing.T) {
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
			jsonFile := filepath.Join(tempDir, tt.name+".json")
			if err := os.WriteFile(jsonFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			hasErr, errMsg, _ := CheckIngestResult(jsonFile)
			if hasErr != tt.hasError {
				t.Errorf("CheckIngestResult() hasError = %v, want %v", hasErr, tt.hasError)
			}
			if tt.wantErrStr && errMsg == "" {
				t.Errorf("CheckIngestResult() expected non-empty error message")
			}
			if !tt.wantErrStr && errMsg != "" {
				t.Errorf("CheckIngestResult() expected empty error message, got %q", errMsg)
			}
		})
	}

	t.Run("Missing file", func(t *testing.T) {
		missingFile := filepath.Join(tempDir, "non_existent.json")
		hasErr, errMsg, err := CheckIngestResult(missingFile)
		if !hasErr {
			t.Errorf("expected hasError = true for missing file")
		}
		if errMsg == "" {
			t.Errorf("expected non-empty errorMsg for missing file")
		}
		if err == nil {
			t.Errorf("expected error object for missing file")
		}
	})
}

func TestCreateUploadRecord(t *testing.T) {
	tempDir := t.TempDir()
	batchDir := filepath.Join(tempDir, "batch_test")
	if err := os.MkdirAll(batchDir, 0755); err != nil {
		t.Fatalf("failed to create batch dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	j := job.Job{
		BaseName:  "sample_archive",
		Batch:     "batch_test",
		Signature: "SIG-999",
	}

	before := time.Now().Add(-1 * time.Second)
	if err := CreateUploadRecord(conf, j); err != nil {
		t.Fatalf("CreateUploadRecord failed: %v", err)
	}
	after := time.Now().Add(1 * time.Second)

	uploadFilePath := filepath.Join(tempDir, "batch_test", "ocfl", "sample_archive.upload.json")
	data, err := os.ReadFile(uploadFilePath)
	if err != nil {
		t.Fatalf("failed to read created upload file: %v", err)
	}

	var info job.UploadInfo
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatalf("failed to unmarshal upload info JSON: %v", err)
	}

	if info.Date == "" {
		t.Fatalf("expected non-empty date in upload info")
	}

	parsedTime, err := time.Parse(time.RFC3339, info.Date)
	if err != nil {
		t.Fatalf("failed to parse date as RFC3339: %v", err)
	}

	if parsedTime.Before(before) || parsedTime.After(after) {
		t.Errorf("parsed time %v out of expected range [%v, %v]", parsedTime, before, after)
	}
}

func TestHandleJobError_Success(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_fail"
	batchDir := filepath.Join(tempDir, batchName)
	inputDir := filepath.Join(batchDir, "incoming")
	errorDir := filepath.Join(batchDir, "error")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Input:   filepath.Join(tempDir, "{batch}", "incoming"),
		Error:   filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	jobDir := filepath.Join(inputDir, "item_err")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}
	infoPath := filepath.Join(jobDir, "info.json")
	if err := os.WriteFile(infoPath, []byte(`{"signature":"SIG-FAIL"}`), 0644); err != nil {
		t.Fatalf("failed to write info.json: %v", err)
	}

	j := job.Job{
		BaseName:  "item_err",
		Batch:     batchName,
		Signature: "SIG-FAIL",
		InfoFile:  infoPath,
	}

	cmdStr := "gocfl create /path/to/item_err.zip /path/to/data -i SIG-FAIL"
	cmdOutput := "Error: invalid metadata schema\nprocess exited with status 1"
	if err := HandleJobError(conf, j, cmdStr, cmdOutput); err != nil {
		t.Fatalf("HandleJobError failed: %v", err)
	}

	// Verify job directory moved from incoming to error
	if _, err := os.Stat(jobDir); !os.IsNotExist(err) {
		t.Errorf("expected source job dir to be removed from incoming, stat returned: %v", err)
	}

	destJobDir := filepath.Join(errorDir, "item_err")
	if fi, err := os.Stat(destJobDir); err != nil || !fi.IsDir() {
		t.Errorf("expected destination job dir in error folder, stat returned: %v", err)
	}

	destInfo := filepath.Join(destJobDir, "info.json")
	if _, err := os.Stat(destInfo); err != nil {
		t.Errorf("expected info.json inside destination job dir, stat returned: %v", err)
	}

	// Verify log file written to error folder with command prepended
	logFile := filepath.Join(errorDir, "item_err.log")
	logContent, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	expectedLog := cmdStr + "\n\n" + cmdOutput
	if string(logContent) != expectedLog {
		t.Errorf("log file content = %q, want %q", string(logContent), expectedLog)
	}
}

func TestHandleJobError_WithCompanionJson(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_comp_err"
	batchDir := filepath.Join(tempDir, batchName)
	inputDir := filepath.Join(batchDir, "incoming")
	errorDir := filepath.Join(batchDir, "error")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Input:   filepath.Join(tempDir, "{batch}", "incoming"),
		Error:   filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	jobDir := filepath.Join(inputDir, "item_comp")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}

	// External companion JSON in incoming directory
	compJsonPath := filepath.Join(inputDir, "item_comp.json")
	compJsonContent := `{"signature":"SIG-COMP","title":"Companion Title"}`
	if err := os.WriteFile(compJsonPath, []byte(compJsonContent), 0644); err != nil {
		t.Fatalf("failed to write companion json: %v", err)
	}

	j := job.Job{
		BaseName:  "item_comp",
		Batch:     batchName,
		Signature: "SIG-COMP",
		InfoFile:  compJsonPath,
	}

	cmdStr := "gocfl create ..."
	cmdOutput := "Error processing"
	if err := HandleJobError(conf, j, cmdStr, cmdOutput); err != nil {
		t.Fatalf("HandleJobError failed: %v", err)
	}

	// Verify job directory moved
	if _, err := os.Stat(jobDir); !os.IsNotExist(err) {
		t.Errorf("expected source job dir to be removed from incoming")
	}
	if _, err := os.Stat(filepath.Join(errorDir, "item_comp")); err != nil {
		t.Errorf("expected job dir in error folder")
	}

	// Verify companion JSON moved from incoming to error
	if _, err := os.Stat(compJsonPath); !os.IsNotExist(err) {
		t.Errorf("expected companion json in incoming to be removed")
	}
	destCompJson := filepath.Join(errorDir, "item_comp.json")
	data, err := os.ReadFile(destCompJson)
	if err != nil {
		t.Fatalf("failed to read companion json in error folder: %v", err)
	}
	if string(data) != compJsonContent {
		t.Errorf("companion json content = %q, want %q", string(data), compJsonContent)
	}
}

func TestHandleJobError_CleansPartialZip(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_partial"
	batchDir := filepath.Join(tempDir, batchName)
	inputDir := filepath.Join(batchDir, "incoming")
	ocflDir := filepath.Join(batchDir, "ocfl")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Input:   filepath.Join(tempDir, "{batch}", "incoming"),
		Error:   filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	jobDir := filepath.Join(inputDir, "item_partial")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}

	// Create partial zip
	partialZip := filepath.Join(ocflDir, "item_partial.zip")
	if err := os.WriteFile(partialZip, []byte("corrupt partial zip"), 0644); err != nil {
		t.Fatalf("failed to create partial zip: %v", err)
	}

	j := job.Job{
		BaseName:  "item_partial",
		Batch:     batchName,
		Signature: "SIG-PARTIAL",
		OcflFile:  partialZip,
	}

	if err := HandleJobError(conf, j, "gocfl create ...", "Creation failed midway"); err != nil {
		t.Fatalf("HandleJobError failed: %v", err)
	}

	// Check that partial zip was deleted
	if _, err := os.Stat(partialZip); !os.IsNotExist(err) {
		t.Errorf("expected partial zip to be removed from ocfl dir, stat returned: %v", err)
	}
}

func TestHandleIngestError_WithResultJSON(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_ingest_err"
	batchDir := filepath.Join(tempDir, batchName)
	inputDir := filepath.Join(batchDir, "incoming")
	errorDir := filepath.Join(batchDir, "error")
	ocflDir := filepath.Join(batchDir, "ocfl")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Input:   filepath.Join(tempDir, "{batch}", "incoming"),
		Error:   filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	jobDir := filepath.Join(inputDir, "item_ingest_err")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}
	infoPath := filepath.Join(jobDir, "info.json")
	if err := os.WriteFile(infoPath, []byte(`{"signature":"SIG-ING-ERR"}`), 0644); err != nil {
		t.Fatalf("failed to write info.json: %v", err)
	}

	// Create OCFL zip
	zipPath := filepath.Join(ocflDir, "item_ingest_err.zip")
	if err := os.WriteFile(zipPath, []byte("ocfl zip content"), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}

	// Create Result JSON with error
	resultJsonPath := filepath.Join(ocflDir, "item_ingest_err_ona.json")
	resultJsonContent := `{"status":"error","error":"remote endpoint unavailable"}`
	if err := os.WriteFile(resultJsonPath, []byte(resultJsonContent), 0644); err != nil {
		t.Fatalf("failed to write result json: %v", err)
	}

	j := job.Job{
		BaseName:  "item_ingest_err",
		Batch:     batchName,
		Signature: "SIG-ING-ERR",
		InfoFile:  infoPath,
		OcflFile:  zipPath,
	}

	cmdStr := "ona ingest --status /path/to/item_ingest_err_ona.json -p /path/to/item_ingest_err.zip"
	cmdOutput := "Upload failed: HTTP 503 Service Unavailable"
	if err := HandleIngestError(conf, j, cmdStr, cmdOutput, resultJsonPath); err != nil {
		t.Fatalf("HandleIngestError failed: %v", err)
	}

	// 1. Verify job directory moved from incoming to error
	if _, err := os.Stat(jobDir); !os.IsNotExist(err) {
		t.Errorf("expected source job dir to be removed from incoming, stat returned: %v", err)
	}
	destJobDir := filepath.Join(errorDir, "item_ingest_err")
	if fi, err := os.Stat(destJobDir); err != nil || !fi.IsDir() {
		t.Errorf("expected destination job dir in error folder, stat returned: %v", err)
	}

	// 2. Verify result JSON moved from ocfl to error
	if _, err := os.Stat(resultJsonPath); !os.IsNotExist(err) {
		t.Errorf("expected result JSON to be removed from ocfl dir, stat returned: %v", err)
	}
	destJsonPath := filepath.Join(errorDir, "item_ingest_err_ona.json")
	destJsonData, err := os.ReadFile(destJsonPath)
	if err != nil {
		t.Fatalf("failed to read destination result json: %v", err)
	}
	if string(destJsonData) != resultJsonContent {
		t.Errorf("dest result json content = %q, want %q", string(destJsonData), resultJsonContent)
	}

	// 3. Verify error log file
	logFile := filepath.Join(errorDir, "item_ingest_err.log")
	logContent, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	expectedLog := cmdStr + "\n\n" + cmdOutput
	if string(logContent) != expectedLog {
		t.Errorf("log file content = %q, want %q", string(logContent), expectedLog)
	}

	// 4. Verify OCFL zip removed
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Errorf("expected ocfl zip to be removed from ocfl dir, stat returned: %v", err)
	}
}

func TestHandleIngestError_WithoutResultJSON(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_no_json"
	batchDir := filepath.Join(tempDir, batchName)
	inputDir := filepath.Join(batchDir, "incoming")
	errorDir := filepath.Join(batchDir, "error")
	ocflDir := filepath.Join(batchDir, "ocfl")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Input:   filepath.Join(tempDir, "{batch}", "incoming"),
		Error:   filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	jobDir := filepath.Join(inputDir, "item_no_json")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}

	zipPath := filepath.Join(ocflDir, "item_no_json.zip")
	if err := os.WriteFile(zipPath, []byte("ocfl zip content"), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}

	j := job.Job{
		BaseName:  "item_no_json",
		Batch:     batchName,
		Signature: "SIG-NO-JSON",
		OcflFile:  zipPath,
	}

	cmdStr := "ona ingest -p item_no_json.zip"
	cmdOutput := "ona executable not found"
	resultJsonPath := filepath.Join(ocflDir, "item_no_json_ona.json")

	if err := HandleIngestError(conf, j, cmdStr, cmdOutput, resultJsonPath); err != nil {
		t.Fatalf("HandleIngestError failed: %v", err)
	}

	// Verify job directory moved and log created even though result JSON was absent
	if _, err := os.Stat(jobDir); !os.IsNotExist(err) {
		t.Errorf("expected source job dir to be removed from incoming")
	}
	destJobDir := filepath.Join(errorDir, "item_no_json")
	if fi, err := os.Stat(destJobDir); err != nil || !fi.IsDir() {
		t.Errorf("expected destination job dir in error folder")
	}

	logFile := filepath.Join(errorDir, "item_no_json.log")
	logContent, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	expectedLog := cmdStr + "\n\n" + cmdOutput
	if string(logContent) != expectedLog {
		t.Errorf("log file content = %q, want %q", string(logContent), expectedLog)
	}

	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Errorf("expected ocfl zip to be removed")
	}
}

func TestHandleJobError_PrependsCommandToLog(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_cmd_log"
	inputDir := filepath.Join(tempDir, batchName, "incoming")
	errorDir := filepath.Join(tempDir, batchName, "error")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches: tempDir,
		Input:   filepath.Join(tempDir, "{batch}", "incoming"),
		Error:   filepath.Join(tempDir, "{batch}", "error"),
		Ocfl:    filepath.Join(tempDir, "{batch}", "ocfl"),
	}

	// Case 1: with non-empty command string
	job1Dir := filepath.Join(inputDir, "job1")
	if err := os.MkdirAll(job1Dir, 0755); err != nil {
		t.Fatalf("failed to create job1 dir: %v", err)
	}
	j1 := job.Job{
		BaseName: "job1",
		Batch:    batchName,
	}
	cmd1 := "gocfl create C:/batch/ocfl/job1.zip C:/batch/incoming/job1/data -i SIG-1"
	out1 := "Fatal error: invalid checksum"
	if err := HandleJobError(conf, j1, cmd1, out1); err != nil {
		t.Fatalf("HandleJobError failed for j1: %v", err)
	}

	log1Path := filepath.Join(errorDir, "job1.log")
	data1, err := os.ReadFile(log1Path)
	if err != nil {
		t.Fatalf("failed to read job1.log: %v", err)
	}
	expected1 := "gocfl create C:/batch/ocfl/job1.zip C:/batch/incoming/job1/data -i SIG-1\n\nFatal error: invalid checksum"
	if string(data1) != expected1 {
		t.Errorf("job1.log content = %q, want %q", string(data1), expected1)
	}

	// Case 2: with empty command string (fallback directly to output)
	job2Dir := filepath.Join(inputDir, "job2")
	if err := os.MkdirAll(job2Dir, 0755); err != nil {
		t.Fatalf("failed to create job2 dir: %v", err)
	}
	j2 := job.Job{
		BaseName: "job2",
		Batch:    batchName,
	}
	out2 := "Plain error without command"
	if err := HandleJobError(conf, j2, "", out2); err != nil {
		t.Fatalf("HandleJobError failed for j2: %v", err)
	}

	log2Path := filepath.Join(errorDir, "job2.log")
	data2, err := os.ReadFile(log2Path)
	if err != nil {
		t.Fatalf("failed to read job2.log: %v", err)
	}
	if string(data2) != out2 {
		t.Errorf("job2.log content = %q, want %q", string(data2), out2)
	}
}

func TestArchiveJob_AllArtifacts(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_archive"
	batchDir := filepath.Join(tempDir, batchName)
	inputDir := filepath.Join(batchDir, "incoming")
	ocflDir := filepath.Join(batchDir, "ocfl")
	reportDir := filepath.Join(batchDir, "report")
	archivedDir := filepath.Join(batchDir, "archived")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		t.Fatalf("failed to create report dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "incoming"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:   filepath.Join(tempDir, "{batch}", "report"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	jobDir := filepath.Join(inputDir, "item_success")
	if err := os.MkdirAll(filepath.Join(jobDir, "data"), 0755); err != nil {
		t.Fatalf("failed to create job data dir: %v", err)
	}
	infoPath := filepath.Join(jobDir, "info.json")
	if err := os.WriteFile(infoPath, []byte(`{"signature":"SIG-OK"}`), 0644); err != nil {
		t.Fatalf("failed to write info.json: %v", err)
	}
	dataPath := filepath.Join(jobDir, "data", "test.bin")
	if err := os.WriteFile(dataPath, []byte("data payload"), 0644); err != nil {
		t.Fatalf("failed to write data: %v", err)
	}

	// OCFL zip
	zipPath := filepath.Join(ocflDir, "item_success.zip")
	if err := os.WriteFile(zipPath, []byte("ocfl zip"), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}

	// Result JSON
	resultJsonPath := filepath.Join(ocflDir, "item_success_ona.json")
	if err := os.WriteFile(resultJsonPath, []byte(`{"status":"success"}`), 0644); err != nil {
		t.Fatalf("failed to write result json: %v", err)
	}

	// Upload JSON
	uploadJsonPath := filepath.Join(ocflDir, "item_success.upload.json")
	if err := os.WriteFile(uploadJsonPath, []byte(`{"date":"2026-10-13T12:00:00Z"}`), 0644); err != nil {
		t.Fatalf("failed to write upload json: %v", err)
	}

	// Report PDF
	reportPath := filepath.Join(reportDir, "item_success.pdf")
	if err := os.WriteFile(reportPath, []byte("%PDF-1.4 report"), 0644); err != nil {
		t.Fatalf("failed to write report: %v", err)
	}

	j := job.Job{
		BaseName:       "item_success",
		Batch:          batchName,
		IncomingFolder: jobDir,
		OcflFile:       zipPath,
		ReportFile:     reportPath,
		UploadFile:     uploadJsonPath,
		Signature:      "SIG-OK",
	}

	if err := ArchiveJob(conf, j, resultJsonPath); err != nil {
		t.Fatalf("ArchiveJob failed: %v", err)
	}

	// 1. Assert incoming directory was moved to archived
	if _, err := os.Stat(jobDir); !os.IsNotExist(err) {
		t.Errorf("expected source job dir in incoming to be removed, stat err: %v", err)
	}
	destJobDir := filepath.Join(archivedDir, "item_success")
	if fi, err := os.Stat(destJobDir); err != nil || !fi.IsDir() {
		t.Errorf("expected archived job dir to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destJobDir, "info.json")); err != nil {
		t.Errorf("expected info.json inside archived job dir")
	}
	if _, err := os.Stat(filepath.Join(destJobDir, "data", "test.bin")); err != nil {
		t.Errorf("expected data/test.bin inside archived job dir")
	}

	// 2. Assert OCFL zip was moved to archived
	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Errorf("expected OCFL zip in ocfl dir to be removed")
	}
	if _, err := os.Stat(filepath.Join(archivedDir, "item_success.zip")); err != nil {
		t.Errorf("expected item_success.zip in archived dir")
	}

	// 3. Assert Result JSON was moved to archived
	if _, err := os.Stat(resultJsonPath); !os.IsNotExist(err) {
		t.Errorf("expected result json in ocfl dir to be removed")
	}
	if _, err := os.Stat(filepath.Join(archivedDir, "item_success_ona.json")); err != nil {
		t.Errorf("expected item_success_ona.json in archived dir")
	}

	// 4. Assert Upload JSON was moved to archived
	if _, err := os.Stat(uploadJsonPath); !os.IsNotExist(err) {
		t.Errorf("expected upload json in ocfl dir to be removed")
	}
	if _, err := os.Stat(filepath.Join(archivedDir, "item_success.upload.json")); err != nil {
		t.Errorf("expected item_success.upload.json in archived dir")
	}

	// 5. Assert PDF report remains in report folder and is NOT in archived
	if _, err := os.Stat(reportPath); err != nil {
		t.Errorf("expected PDF report to remain in report folder, got err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(archivedDir, "item_success.pdf")); !os.IsNotExist(err) {
		t.Errorf("expected PDF report NOT to be in archived dir")
	}
}

func TestArchiveJob_WithoutOptionalFiles(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_archive_minimal"
	batchDir := filepath.Join(tempDir, batchName)
	inputDir := filepath.Join(batchDir, "incoming")
	ocflDir := filepath.Join(batchDir, "ocfl")
	archivedDir := filepath.Join(batchDir, "archived")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "incoming"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	jobDir := filepath.Join(inputDir, "item_min")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}

	zipPath := filepath.Join(ocflDir, "item_min.zip")
	if err := os.WriteFile(zipPath, []byte("ocfl zip"), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}

	j := job.Job{
		BaseName:  "item_min",
		Batch:     batchName,
		Signature: "SIG-MIN",
	}

	// Pass empty resultJsonPath (no result json created)
	if err := ArchiveJob(conf, j, ""); err != nil {
		t.Fatalf("ArchiveJob failed on minimal job: %v", err)
	}

	if _, err := os.Stat(jobDir); !os.IsNotExist(err) {
		t.Errorf("expected source job dir to be removed")
	}
	if _, err := os.Stat(filepath.Join(archivedDir, "item_min")); err != nil {
		t.Errorf("expected archived job dir to exist")
	}
	if _, err := os.Stat(filepath.Join(archivedDir, "item_min.zip")); err != nil {
		t.Errorf("expected item_min.zip in archived dir")
	}
}

func TestArchiveJob_WithCompanionJsonAndOnaJson(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_archive_comp"
	batchDir := filepath.Join(tempDir, batchName)
	inputDir := filepath.Join(batchDir, "incoming")
	ocflDir := filepath.Join(batchDir, "ocfl")
	archivedDir := filepath.Join(batchDir, "archived")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}

	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "incoming"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	jobDir := filepath.Join(inputDir, "item_comp_arch")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}

	// External companion JSON in incoming
	compJsonPath := filepath.Join(inputDir, "item_comp_arch.json")
	compContent := `{"signature":"SIG-ARCH-COMP","title":"Metadata Title"}`
	if err := os.WriteFile(compJsonPath, []byte(compContent), 0644); err != nil {
		t.Fatalf("failed to write comp json: %v", err)
	}

	// OCFL zip
	zipPath := filepath.Join(ocflDir, "item_comp_arch.zip")
	if err := os.WriteFile(zipPath, []byte("ocfl zip content"), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}

	// ONA result JSON
	onaJsonPath := filepath.Join(ocflDir, "item_comp_arch_ona.json")
	onaContent := `{"id":"ona-arch-789","status":"ok"}`
	if err := os.WriteFile(onaJsonPath, []byte(onaContent), 0644); err != nil {
		t.Fatalf("failed to write ona json: %v", err)
	}

	j := job.Job{
		BaseName:  "item_comp_arch",
		Batch:     batchName,
		Signature: "SIG-ARCH-COMP",
		InfoFile:  compJsonPath,
		OcflFile:  zipPath,
	}

	if err := ArchiveJob(conf, j, onaJsonPath); err != nil {
		t.Fatalf("ArchiveJob failed: %v", err)
	}

	// Assert job directory moved
	if _, err := os.Stat(jobDir); !os.IsNotExist(err) {
		t.Errorf("expected source job dir in incoming to be removed")
	}
	if _, err := os.Stat(filepath.Join(archivedDir, "item_comp_arch")); err != nil {
		t.Errorf("expected archived job dir to exist")
	}

	// Assert companion JSON moved to archived
	if _, err := os.Stat(compJsonPath); !os.IsNotExist(err) {
		t.Errorf("expected companion json in incoming to be removed")
	}
	destComp := filepath.Join(archivedDir, "item_comp_arch.json")
	compData, err := os.ReadFile(destComp)
	if err != nil {
		t.Fatalf("failed to read archived companion json: %v", err)
	}
	if string(compData) != compContent {
		t.Errorf("archived companion json content = %q, want %q", string(compData), compContent)
	}

	// Assert ONA JSON moved to archived
	if _, err := os.Stat(onaJsonPath); !os.IsNotExist(err) {
		t.Errorf("expected ona json in ocfl to be removed")
	}
	destOna := filepath.Join(archivedDir, "item_comp_arch_ona.json")
	onaData, err := os.ReadFile(destOna)
	if err != nil {
		t.Fatalf("failed to read archived ona json: %v", err)
	}
	if string(onaData) != onaContent {
		t.Errorf("archived ona json content = %q, want %q", string(onaData), compContent)
	}

	// Assert OCFL zip moved to archived
	if _, err := os.Stat(filepath.Join(archivedDir, "item_comp_arch.zip")); err != nil {
		t.Errorf("expected zip in archived dir")
	}
}

func TestStatusOCFLCmd(t *testing.T) {
	conf := &config.WBConfig{
		Ona: config.ONAConfig{
			Config: "C:/path/to/ona.yml",
		},
	}

	cmd, err := StatusOCFLCmd(conf, "pkg-12345")
	if err != nil {
		t.Fatalf("StatusOCFLCmd failed: %v", err)
	}

	expectedArgs := []string{"ona", "status", "-i", "pkg-12345", "-c", "C:/path/to/ona.yml"}
	if len(cmd.Args) != len(expectedArgs) {
		t.Fatalf("cmd.Args len = %d, want %d: %v", len(cmd.Args), len(expectedArgs), cmd.Args)
	}
	for i := range expectedArgs {
		if cmd.Args[i] != expectedArgs[i] {
			t.Errorf("cmd.Args[%d] = %s, want %s", i, cmd.Args[i], expectedArgs[i])
		}
	}
}

func TestStatusOCFLCmd_WithAndWithoutConfig(t *testing.T) {
	// Without config
	confNoConfig := &config.WBConfig{}
	cmd, err := StatusOCFLCmd(confNoConfig, "pkg-67890")
	if err != nil {
		t.Fatalf("StatusOCFLCmd without config failed: %v", err)
	}
	expectedArgsNoConfig := []string{"ona", "status", "-i", "pkg-67890"}
	if len(cmd.Args) != len(expectedArgsNoConfig) {
		t.Fatalf("cmd.Args len = %d, want %d", len(cmd.Args), len(expectedArgsNoConfig))
	}
	for i := range expectedArgsNoConfig {
		if cmd.Args[i] != expectedArgsNoConfig[i] {
			t.Errorf("cmd.Args[%d] = %s, want %s", i, cmd.Args[i], expectedArgsNoConfig[i])
		}
	}

	// With nil config
	cmdNil, err := StatusOCFLCmd(nil, "pkg-nil")
	if err != nil {
		t.Fatalf("StatusOCFLCmd with nil config failed: %v", err)
	}
	if len(cmdNil.Args) != 4 || cmdNil.Args[2] != "-i" || cmdNil.Args[3] != "pkg-nil" {
		t.Errorf("unexpected cmdArgs for nil config: %v", cmdNil.Args)
	}

	// Empty ID should return error
	_, errEmpty := StatusOCFLCmd(confNoConfig, "")
	if errEmpty == nil {
		t.Errorf("expected error for empty id, got nil")
	}

	_, errWhitespace := StatusOCFLCmd(confNoConfig, "   ")
	if errWhitespace == nil {
		t.Errorf("expected error for whitespace id, got nil")
	}
}

func TestExtractIngestID_Runner(t *testing.T) {
	tempDir := t.TempDir()

	validJson := filepath.Join(tempDir, "valid.json")
	if err := os.WriteFile(validJson, []byte(`{"id":"ona-valid-123","status":"ok"}`), 0644); err != nil {
		t.Fatalf("failed to write valid.json: %v", err)
	}

	id, err := ExtractIngestID(validJson)
	if err != nil {
		t.Fatalf("ExtractIngestID failed: %v", err)
	}
	if id != "ona-valid-123" {
		t.Errorf("ExtractIngestID = %q, want %q", id, "ona-valid-123")
	}

	missingJson := filepath.Join(tempDir, "missing.json")
	if err := os.WriteFile(missingJson, []byte(`{"status":"ok"}`), 0644); err != nil {
		t.Fatalf("failed to write missing.json: %v", err)
	}
	_, errMissing := ExtractIngestID(missingJson)
	if errMissing == nil {
		t.Errorf("expected error for json without id, got nil")
	}

	_, errNotFound := ExtractIngestID(filepath.Join(tempDir, "nonexistent.json"))
	if errNotFound == nil {
		t.Errorf("expected error for non-existent file, got nil")
	}
}
