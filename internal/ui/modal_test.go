package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/ocfl-archive/workbench/internal/job"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

func TestRunCommandStream_Echo(t *testing.T) {
	modalText := tview.NewTextView().SetDynamicColors(true)

	var cmd *exec.Cmd
	if os.Getenv("OS") == "Windows_NT" {
		cmd = exec.Command("cmd.exe", "/c", "echo", "test-output-streaming")
	} else {
		cmd = exec.Command("echo", "test-output-streaming")
	}

	exitCode, output := RunCommandStream(nil, modalText, cmd)
	if exitCode != 0 {
		t.Fatalf("expected exitCode 0, got %d (output: %s)", exitCode, output)
	}
	if !strings.Contains(output, "test-output-streaming") {
		t.Errorf("expected output to contain 'test-output-streaming', got %q", output)
	}
}

func TestRunBatchInModal_EmptyOrFinishedBatch(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_empty"

	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "incoming"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:   filepath.Join(tempDir, "{batch}", "report"),
		Error:    filepath.Join(tempDir, "{batch}", "error"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	// Create archived job and error job
	archivedDir := filepath.Join(conf.GetArchivedFolder(batchName), "job_archived")
	if err := os.MkdirAll(archivedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(archivedDir, "info.json"), []byte(`{"signature":"SIG-ARCH","title":"Archived"}`), 0644); err != nil {
		t.Fatal(err)
	}

	errorDir := filepath.Join(conf.GetErrorFolder(batchName), "job_error")
	if err := os.MkdirAll(errorDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(errorDir, "info.json"), []byte(`{"signature":"SIG-ERR","title":"Error"}`), 0644); err != nil {
		t.Fatal(err)
	}

	pages := tview.NewPages()
	logger := zerolog.Nop()

	RunBatchInModal(nil, pages, conf, batchName, logger, func() {})

	// Wait for execModal to be added to pages
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if pages.HasPage("execModal") {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if !pages.HasPage("execModal") {
		t.Errorf("expected pages to contain 'execModal'")
	}
}

func TestRunBatchInModal_MissingFolder(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_nonexistent"

	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "incoming"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:   filepath.Join(tempDir, "{batch}", "report"),
		Error:    filepath.Join(tempDir, "{batch}", "error"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	pages := tview.NewPages()
	logger := zerolog.Nop()

	RunBatchInModal(nil, pages, conf, batchName, logger, func() {})

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if pages.HasPage("execModal") {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if !pages.HasPage("execModal") {
		t.Errorf("expected pages to contain 'execModal'")
	}
}

func TestUI_SetupUI_IncludesBatchButton(t *testing.T) {
	tempDir := t.TempDir()
	batchName := "batch_test"

	conf := &config.WBConfig{
		Batches:  tempDir,
		Input:    filepath.Join(tempDir, "{batch}", "incoming"),
		Ocfl:     filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:   filepath.Join(tempDir, "{batch}", "report"),
		Error:    filepath.Join(tempDir, "{batch}", "error"),
		Archived: filepath.Join(tempDir, "{batch}", "archived"),
	}

	incomingJobDir := filepath.Join(conf.GetInputFolder(batchName), "job_1")
	if err := os.MkdirAll(incomingJobDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(incomingJobDir, "info.json"), []byte(`{"signature":"SIG-1","title":"Test Title"}`), 0644); err != nil {
		t.Fatal(err)
	}

	app := tview.NewApplication()
	logger := zerolog.Nop()

	bannerView, pages := SetupUI(app, conf, logger, strings.NewReader(""))
	if bannerView == nil || pages == nil {
		t.Fatal("SetupUI returned nil bannerView or pages")
	}

	if !pages.HasPage("main") {
		t.Errorf("expected pages to contain 'main' page")
	}

	// Verify ActionRunBatch label is as expected
	if got := job.ActionRunBatch.Label(); got != "Batch durchführen" {
		t.Errorf("ActionRunBatch.Label() = %q, want 'Batch durchführen'", got)
	}
}
