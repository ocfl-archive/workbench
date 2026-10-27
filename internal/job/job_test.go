package job

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ocfl-archive/workbench/internal/config"
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
	jobs, err := GetSignatures(batchName, incomingDir, ocflDir, reportDir, "", "", logger)
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
	if j.Status != StatusIncoming {
		t.Errorf("job.Status = %s, want %s", j.Status, StatusIncoming)
	}
	if j.Signature != "SIG-001" {
		t.Errorf("job.Signature = %s, want SIG-001", j.Signature)
	}
	if j.Title != "Test Job 1" {
		t.Errorf("job.Title = %s, want Test Job 1", j.Title)
	}
	if j.IncomingFolder != job1Dir {
		t.Errorf("job.IncomingFolder = %s, want %s", j.IncomingFolder, job1Dir)
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

func TestGetSignaturesWithIngestResultJson(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_ingest"
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

	jobDir := filepath.Join(incomingDir, "job2")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job2 dir: %v", err)
	}
	infoContent := `{"signature":"SIG-002","title":"Test Job 2"}`
	if err := os.WriteFile(filepath.Join(jobDir, "info.json"), []byte(infoContent), 0644); err != nil {
		t.Fatalf("failed to write job2 info: %v", err)
	}

	// Create result JSON (e.g. job2.json)
	resultJsonPath := filepath.Join(ocflDir, "job2.json")
	if err := os.WriteFile(resultJsonPath, []byte(`{"status":"ok"}`), 0644); err != nil {
		t.Fatalf("failed to write job2.json: %v", err)
	}

	logger := zerolog.Nop()
	jobs, err := GetSignatures(batchName, incomingDir, ocflDir, reportDir, "", "", logger)
	if err != nil {
		t.Fatalf("GetSignatures returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	if jobs[0].UploadFile != resultJsonPath {
		t.Errorf("job.UploadFile = %q, want %q", jobs[0].UploadFile, resultJsonPath)
	}
}

func TestGetSignaturesSiblingInfoJson(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_sibling"
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

	jobDir := filepath.Join(incomingDir, "job_sibling")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job_sibling dir: %v", err)
	}

	// Sibling JSON beside job_sibling/ folder: incoming/job_sibling.json
	siblingJsonPath := filepath.Join(incomingDir, "job_sibling.json")
	siblingContent := `{"signature":"SIG-SIBLING","title":"Sibling Metadata Job"}`
	if err := os.WriteFile(siblingJsonPath, []byte(siblingContent), 0644); err != nil {
		t.Fatalf("failed to write sibling json: %v", err)
	}

	logger := zerolog.Nop()
	jobs, err := GetSignatures(batchName, incomingDir, ocflDir, reportDir, "", "", logger)
	if err != nil {
		t.Fatalf("GetSignatures returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	j := jobs[0]
	if j.BaseName != "job_sibling" {
		t.Errorf("job.BaseName = %s, want job_sibling", j.BaseName)
	}
	if j.IncomingFolder != jobDir {
		t.Errorf("job.IncomingFolder = %s, want %s", j.IncomingFolder, jobDir)
	}
	if j.InfoFile != siblingJsonPath {
		t.Errorf("job.InfoFile = %s, want %s", j.InfoFile, siblingJsonPath)
	}
	if j.Signature != "SIG-SIBLING" {
		t.Errorf("job.Signature = %s, want SIG-SIBLING", j.Signature)
	}
	if j.Title != "Sibling Metadata Job" {
		t.Errorf("job.Title = %s, want Sibling Metadata Job", j.Title)
	}
}

func TestGetSignaturesChildNamedJson(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_child_named"
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

	jobDir := filepath.Join(incomingDir, "job_named")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job_named dir: %v", err)
	}

	// Child JSON inside job_named/ folder: incoming/job_named/job_named.json
	childNamedJsonPath := filepath.Join(jobDir, "job_named.json")
	childContent := `{"signature":"SIG-CHILD-NAMED","title":"Child Named Metadata Job"}`
	if err := os.WriteFile(childNamedJsonPath, []byte(childContent), 0644); err != nil {
		t.Fatalf("failed to write child named json: %v", err)
	}

	logger := zerolog.Nop()
	jobs, err := GetSignatures(batchName, incomingDir, ocflDir, reportDir, "", "", logger)
	if err != nil {
		t.Fatalf("GetSignatures returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	j := jobs[0]
	if j.BaseName != "job_named" {
		t.Errorf("job.BaseName = %s, want job_named", j.BaseName)
	}
	if j.IncomingFolder != jobDir {
		t.Errorf("job.IncomingFolder = %s, want %s", j.IncomingFolder, jobDir)
	}
	if j.InfoFile != childNamedJsonPath {
		t.Errorf("job.InfoFile = %s, want %s", j.InfoFile, childNamedJsonPath)
	}
	if j.Signature != "SIG-CHILD-NAMED" {
		t.Errorf("job.Signature = %s, want SIG-CHILD-NAMED", j.Signature)
	}
	if j.Title != "Child Named Metadata Job" {
		t.Errorf("job.Title = %s, want Child Named Metadata Job", j.Title)
	}
}

func TestGetSignatures_WithErrorAndArchivedFolders(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_multi_status"
	incomingDir := filepath.Join(tempDir, batchName, "incoming")
	errorDir := filepath.Join(tempDir, batchName, "error")
	archivedDir := filepath.Join(tempDir, batchName, "archived")
	ocflDir := filepath.Join(tempDir, batchName, "ocfl")
	reportDir := filepath.Join(tempDir, batchName, "report")

	if err := os.MkdirAll(incomingDir, 0755); err != nil {
		t.Fatalf("failed to create incoming dir: %v", err)
	}
	if err := os.MkdirAll(errorDir, 0755); err != nil {
		t.Fatalf("failed to create error dir: %v", err)
	}
	if err := os.MkdirAll(archivedDir, 0755); err != nil {
		t.Fatalf("failed to create archived dir: %v", err)
	}
	if err := os.MkdirAll(ocflDir, 0755); err != nil {
		t.Fatalf("failed to create ocfl dir: %v", err)
	}
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		t.Fatalf("failed to create report dir: %v", err)
	}

	// 1. Incoming job
	incJobDir := filepath.Join(incomingDir, "job_inc")
	if err := os.MkdirAll(filepath.Join(incJobDir, "data"), 0755); err != nil {
		t.Fatalf("failed to create inc job data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(incJobDir, "info.json"), []byte(`{"signature":"SIG-INC","title":"Incoming Job"}`), 0644); err != nil {
		t.Fatalf("failed to write inc job info: %v", err)
	}

	// 2. Error job
	errJobDir := filepath.Join(errorDir, "job_err")
	if err := os.MkdirAll(filepath.Join(errJobDir, "data"), 0755); err != nil {
		t.Fatalf("failed to create err job data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(errJobDir, "info.json"), []byte(`{"signature":"SIG-ERR","title":"Error Job"}`), 0644); err != nil {
		t.Fatalf("failed to write err job info: %v", err)
	}
	errLogPath := filepath.Join(errorDir, "job_err.log")
	if err := os.WriteFile(errLogPath, []byte("Command failed with error exit 1"), 0644); err != nil {
		t.Fatalf("failed to write error log: %v", err)
	}

	// 3. Archived job
	archJobDir := filepath.Join(archivedDir, "job_arch")
	if err := os.MkdirAll(filepath.Join(archJobDir, "data"), 0755); err != nil {
		t.Fatalf("failed to create arch job data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(archJobDir, "info.json"), []byte(`{"signature":"SIG-ARCH","title":"Archived Job"}`), 0644); err != nil {
		t.Fatalf("failed to write arch job info: %v", err)
	}
	archZipPath := filepath.Join(archivedDir, "job_arch.zip")
	if err := os.WriteFile(archZipPath, []byte("archived-zip-content"), 0644); err != nil {
		t.Fatalf("failed to write arch zip: %v", err)
	}
	archJsonPath := filepath.Join(archivedDir, "job_arch_ona.json")
	if err := os.WriteFile(archJsonPath, []byte(`{"id":"ona-item-999","status":"success"}`), 0644); err != nil {
		t.Fatalf("failed to write arch json: %v", err)
	}

	logger := zerolog.Nop()
	jobs, err := GetSignatures(batchName, incomingDir, ocflDir, reportDir, errorDir, archivedDir, logger)
	if err != nil {
		t.Fatalf("GetSignatures returned error: %v", err)
	}

	if len(jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(jobs))
	}

	var incJob, errJob, archJob *Job
	for i := range jobs {
		switch jobs[i].Status {
		case StatusIncoming:
			incJob = &jobs[i]
		case StatusError:
			errJob = &jobs[i]
		case StatusArchived:
			archJob = &jobs[i]
		}
	}

	if incJob == nil {
		t.Fatalf("missing incoming job")
	}
	if incJob.Signature != "SIG-INC" {
		t.Errorf("incJob.Signature = %s, want SIG-INC", incJob.Signature)
	}

	if errJob == nil {
		t.Fatalf("missing error job")
	}
	if errJob.Signature != "SIG-ERR" {
		t.Errorf("errJob.Signature = %s, want SIG-ERR", errJob.Signature)
	}
	if errJob.ErrorFile != errLogPath {
		t.Errorf("errJob.ErrorFile = %s, want %s", errJob.ErrorFile, errLogPath)
	}

	if archJob == nil {
		t.Fatalf("missing archived job")
	}
	if archJob.Signature != "SIG-ARCH" {
		t.Errorf("archJob.Signature = %s, want SIG-ARCH", archJob.Signature)
	}
	if archJob.IngestID != "ona-item-999" {
		t.Errorf("archJob.IngestID = %s, want %s", archJob.IngestID, "ona-item-999")
	}
	if archJob.OcflFile != archZipPath {
		t.Errorf("archJob.OcflFile = %s, want %s", archJob.OcflFile, archZipPath)
	}
}

func TestGetSignatures_ArchivedWithCompanionMetadataAndOnaJson(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_arch_comp"
	archivedDir := filepath.Join(tempDir, batchName, "archived")

	if err := os.MkdirAll(archivedDir, 0755); err != nil {
		t.Fatalf("failed to create archived dir: %v", err)
	}

	jobDir := filepath.Join(archivedDir, "job_arch_meta")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job dir: %v", err)
	}

	// Companion metadata file in archived directory
	metaPath := filepath.Join(archivedDir, "job_arch_meta.json")
	if err := os.WriteFile(metaPath, []byte(`{"signature":"SIG-COMP-ARCH","title":"Companion Archived Job"}`), 0644); err != nil {
		t.Fatalf("failed to write meta json: %v", err)
	}

	// Ingest status file in archived directory
	onaJsonPath := filepath.Join(archivedDir, "job_arch_meta_ona.json")
	if err := os.WriteFile(onaJsonPath, []byte(`{"id":"ona-ingest-456","status":"success"}`), 0644); err != nil {
		t.Fatalf("failed to write ona json: %v", err)
	}

	// Zip container
	zipPath := filepath.Join(archivedDir, "job_arch_meta.zip")
	if err := os.WriteFile(zipPath, []byte("zip-data"), 0644); err != nil {
		t.Fatalf("failed to write zip: %v", err)
	}

	logger := zerolog.Nop()
	jobs, err := GetSignatures(batchName, "", "", "", "", archivedDir, logger)
	if err != nil {
		t.Fatalf("GetSignatures returned error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	j := jobs[0]
	if j.Status != StatusArchived {
		t.Errorf("job.Status = %s, want %s", j.Status, StatusArchived)
	}
	if j.Signature != "SIG-COMP-ARCH" {
		t.Errorf("job.Signature = %s, want SIG-COMP-ARCH", j.Signature)
	}
	if j.Title != "Companion Archived Job" {
		t.Errorf("job.Title = %s, want Companion Archived Job", j.Title)
	}
	if j.InfoFile != metaPath {
		t.Errorf("job.InfoFile = %s, want %s", j.InfoFile, metaPath)
	}
	if j.IngestID != "ona-ingest-456" {
		t.Errorf("job.IngestID = %s, want ona-ingest-456", j.IngestID)
	}
	if j.OcflFile != zipPath {
		t.Errorf("job.OcflFile = %s, want %s", j.OcflFile, zipPath)
	}
}

func TestGetSignatures_NonExistentFolders(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_nonexistent"
	incomingDir := filepath.Join(tempDir, batchName, "incoming")
	errorDir := filepath.Join(tempDir, batchName, "error")       // does not exist
	archivedDir := filepath.Join(tempDir, batchName, "archived") // does not exist
	ocflDir := filepath.Join(tempDir, batchName, "ocfl")
	reportDir := filepath.Join(tempDir, batchName, "report")

	if err := os.MkdirAll(incomingDir, 0755); err != nil {
		t.Fatalf("failed to create incoming dir: %v", err)
	}

	jobDir := filepath.Join(incomingDir, "job1")
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		t.Fatalf("failed to create job1 dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(jobDir, "info.json"), []byte(`{"signature":"SIG-1","title":"Job 1"}`), 0644); err != nil {
		t.Fatalf("failed to write job1 info: %v", err)
	}

	logger := zerolog.Nop()
	jobs, err := GetSignatures(batchName, incomingDir, ocflDir, reportDir, errorDir, archivedDir, logger)
	if err != nil {
		t.Fatalf("GetSignatures should succeed with missing error/archived dirs, got error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Status != StatusIncoming {
		t.Errorf("job.Status = %s, want %s", jobs[0].Status, StatusIncoming)
	}
}

func TestExtractIngestID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "string id",
			input:    `{"id": "ona-abc-123", "status": "done"}`,
			expected: "ona-abc-123",
		},
		{
			name:     "integer id",
			input:    `{"id": 4567, "status": "done"}`,
			expected: "4567",
		},
		{
			name:     "float/numeric id",
			input:    `{"id": 7890.0, "status": "done"}`,
			expected: "7890",
		},
		{
			name:     "missing id",
			input:    `{"status": "done"}`,
			expected: "",
		},
		{
			name:     "null id",
			input:    `{"id": null}`,
			expected: "",
		},
		{
			name:     "empty id",
			input:    `{"id": ""}`,
			expected: "",
		},
		{
			name:     "malformed json",
			input:    `{invalid`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractIngestID([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("ExtractIngestID() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestJobAvailableActions(t *testing.T) {
	tests := []struct {
		name     string
		job      Job
		expected []ActionType
	}{
		{
			name: "error status has no actions",
			job: Job{
				Status:    StatusError,
				Signature: "SIG-1",
			},
			expected: nil,
		},
		{
			name: "missing signature has no actions",
			job: Job{
				Status:    StatusIncoming,
				Signature: "",
			},
			expected: nil,
		},
		{
			name: "archived without ingest id has no actions",
			job: Job{
				Status:    StatusArchived,
				Signature: "SIG-1",
				IngestID:  "",
			},
			expected: nil,
		},
		{
			name: "archived with ingest id has check status action",
			job: Job{
				Status:    StatusArchived,
				Signature: "SIG-1",
				IngestID:  "12345",
			},
			expected: []ActionType{ActionCheckStatus},
		},
		{
			name: "incoming without ocfl file has create ocfl action",
			job: Job{
				Status:    StatusIncoming,
				Signature: "SIG-1",
				OcflFile:  "",
			},
			expected: []ActionType{ActionCreateOCFL},
		},
		{
			name: "incoming with ocfl file but without report has report, validate and ingest",
			job: Job{
				Status:     StatusIncoming,
				Signature:  "SIG-1",
				OcflFile:   "/path/to/item.zip",
				ReportFile: "",
			},
			expected: []ActionType{ActionCreateReport, ActionValidate, ActionIngest},
		},
		{
			name: "incoming with ocfl and report has validate and ingest",
			job: Job{
				Status:     StatusIncoming,
				Signature:  "SIG-1",
				OcflFile:   "/path/to/item.zip",
				ReportFile: "/path/to/item.pdf",
			},
			expected: []ActionType{ActionValidate, ActionIngest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.AvailableActions()
			if len(got) != len(tt.expected) {
				t.Fatalf("AvailableActions() returned %d actions, want %d: %v", len(got), len(tt.expected), got)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("AvailableActions()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestActionTypeLabels(t *testing.T) {
	cases := map[ActionType]string{
		ActionCreateOCFL:    "OCFL erstellen",
		ActionCreateReport:  "Report erstellen",
		ActionValidate:      "Validate",
		ActionIngest:        "Ingest",
		ActionCheckStatus:   "Status abfragen",
		ActionType("other"): "other",
	}

	for act, expectedLabel := range cases {
		if got := act.Label(); got != expectedLabel {
			t.Errorf("ActionType(%s).Label() = %q, want %q", act, got, expectedLabel)
		}
	}
}

func TestJobExpectedPaths(t *testing.T) {
	tempDir := t.TempDir()
	conf := &config.WBConfig{
		Batches:  tempDir,
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:   filepath.Join(tempDir, "{batch}", "report"),
		Error:    filepath.Join(tempDir, "{batch}", "error"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	j := Job{
		BaseName:  "item_123",
		Batch:     "batch_a",
		Signature: "SIG-123",
	}

	if j.ExpectedZipName() != "item_123.zip" {
		t.Errorf("ExpectedZipName() = %q, want item_123.zip", j.ExpectedZipName())
	}
	if j.ExpectedReportName() != "item_123.pdf" {
		t.Errorf("ExpectedReportName() = %q, want item_123.pdf", j.ExpectedReportName())
	}
	if j.ExpectedResultJsonName() != "item_123_ona.json" {
		t.Errorf("ExpectedResultJsonName() = %q, want item_123_ona.json", j.ExpectedResultJsonName())
	}
	if j.ExpectedUploadJsonName() != "item_123.upload.json" {
		t.Errorf("ExpectedUploadJsonName() = %q, want item_123.upload.json", j.ExpectedUploadJsonName())
	}
	if j.ExpectedErrorLogName() != "item_123.log" {
		t.Errorf("ExpectedErrorLogName() = %q, want item_123.log", j.ExpectedErrorLogName())
	}

	expectedZipPath := filepath.Join(tempDir, "batch_a", "ocfl", "item_123.zip")
	if got := j.ExpectedZipPath(conf); got != expectedZipPath {
		t.Errorf("ExpectedZipPath() = %q, want %q", got, expectedZipPath)
	}

	expectedReportPath := filepath.Join(tempDir, "batch_a", "report", "item_123.pdf")
	if got := j.ExpectedReportPath(conf); got != expectedReportPath {
		t.Errorf("ExpectedReportPath() = %q, want %q", got, expectedReportPath)
	}

	expectedResultJsonPath := filepath.Join(tempDir, "batch_a", "ocfl", "item_123_ona.json")
	if got := j.ExpectedResultJsonPath(conf); got != expectedResultJsonPath {
		t.Errorf("ExpectedResultJsonPath() = %q, want %q", got, expectedResultJsonPath)
	}

	expectedUploadPath := filepath.Join(tempDir, "batch_a", "ocfl", "item_123.upload.json")
	if got := j.ExpectedUploadPath(conf); got != expectedUploadPath {
		t.Errorf("ExpectedUploadPath() = %q, want %q", got, expectedUploadPath)
	}

	expectedErrorLogPath := filepath.Join(tempDir, "batch_a", "error", "item_123.log")
	if got := j.ExpectedErrorLogPath(conf); got != expectedErrorLogPath {
		t.Errorf("ExpectedErrorLogPath() = %q, want %q", got, expectedErrorLogPath)
	}

	// Nil config handling
	if got := j.ExpectedZipPath(nil); got != "" {
		t.Errorf("ExpectedZipPath(nil) = %q, want empty", got)
	}
	if got := j.ExpectedReportPath(nil); got != "" {
		t.Errorf("ExpectedReportPath(nil) = %q, want empty", got)
	}
	if got := j.ExpectedResultJsonPath(nil); got != "" {
		t.Errorf("ExpectedResultJsonPath(nil) = %q, want empty", got)
	}
	if got := j.ExpectedUploadPath(nil); got != "" {
		t.Errorf("ExpectedUploadPath(nil) = %q, want empty", got)
	}
	if got := j.ExpectedErrorLogPath(nil); got != "" {
		t.Errorf("ExpectedErrorLogPath(nil) = %q, want empty", got)
	}
}
