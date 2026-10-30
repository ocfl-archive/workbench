package ui

import (
	"strings"
	"testing"

	"github.com/ocfl-archive/workbench/internal/job"
	"github.com/rivo/tview"
)

func TestShowDetail(t *testing.T) {
	tv := tview.NewTextView().SetDynamicColors(true)

	j := job.Job{
		BaseName:       "job_1",
		Batch:          "batch_a",
		IncomingFolder: "C:/temp/zhbluzern/raphael_upload/batch_a/incoming/job_1",
		InfoFile:       "C:/temp/zhbluzern/raphael_upload/batch_a/incoming/job_1/info.json",
		DataFolder:     "C:/temp/zhbluzern/raphael_upload/batch_a/incoming/job_1/data",
		MetadataFolder: "C:/temp/zhbluzern/raphael_upload/batch_a/incoming/job_1/metadata",
		OcflFile:       "C:/temp/zhbluzern/raphael_upload/batch_a/ocfl/job_1.zip",
		ChecksumFile:   "C:/temp/zhbluzern/raphael_upload/batch_a/ocfl/job_1.zip.sha512",
		ReportFile:     "C:/temp/zhbluzern/raphael_upload/batch_a/report/job_1.pdf",
		UploadFile:     "C:/temp/zhbluzern/raphael_upload/batch_a/ocfl/job_1.upload.json",
		Signature:      "SIG-1234",
		Title:          "Sample Manuscript",
	}

	ShowDetail(j, tv)
	text := tv.GetText(true)

	expectedSubstrings := []string{
		"Signature: SIG-1234",
		"Title:     Sample Manuscript",
		"Incoming Folder: C:/temp/zhbluzern/raphael_upload/batch_a/incoming/job_1",
		"Info File:       C:/temp/zhbluzern/raphael_upload/batch_a/incoming/job_1/info.json",
		"Data Folder:     C:/temp/zhbluzern/raphael_upload/batch_a/incoming/job_1/data",
		"Metadata Folder: C:/temp/zhbluzern/raphael_upload/batch_a/incoming/job_1/metadata",
		"Report File:     C:/temp/zhbluzern/raphael_upload/batch_a/report/job_1.pdf",
		"OCFL File:       C:/temp/zhbluzern/raphael_upload/batch_a/ocfl/job_1.zip",
		"Checksum File:   C:/temp/zhbluzern/raphael_upload/batch_a/ocfl/job_1.zip.sha512",
		"Upload File:     C:/temp/zhbluzern/raphael_upload/batch_a/ocfl/job_1.upload.json",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(text, sub) {
			t.Errorf("expected detail view text to contain %q, but got:\n%s", sub, text)
		}
	}
}

func TestShowDetailWithoutIncomingFolder(t *testing.T) {
	tv := tview.NewTextView().SetDynamicColors(true)

	j := job.Job{
		BaseName:  "job_2",
		Batch:     "batch_b",
		InfoFile:  "C:/temp/info.json",
		Signature: "SIG-5678",
		Title:     "Another Document",
	}

	ShowDetail(j, tv)
	text := tv.GetText(true)

	if strings.Contains(text, "Incoming Folder:") {
		t.Errorf("did not expect Incoming Folder in output when unset, got:\n%s", text)
	}
	if !strings.Contains(text, "Info File:       C:/temp/info.json") {
		t.Errorf("expected Info File in output, got:\n%s", text)
	}
}

func TestShowDetail_StatusAndIngestID(t *testing.T) {
	tv := tview.NewTextView().SetDynamicColors(true)

	j := job.Job{
		BaseName:  "job_arch",
		Batch:     "batch_arch",
		Status:    job.StatusArchived,
		IngestID:  "ona-obj-9876",
		Signature: "SIG-ARCH-1",
		Title:     "Archived Record",
	}

	ShowDetail(j, tv)
	text := tv.GetText(true)

	if !strings.Contains(text, "Status:    archived") {
		t.Errorf("expected Status in output, got:\n%s", text)
	}
	if !strings.Contains(text, "Ingest ID: ona-obj-9876") {
		t.Errorf("expected Ingest ID in output, got:\n%s", text)
	}
}

func TestShowDetail_ErrorJob(t *testing.T) {
	tv := tview.NewTextView().SetDynamicColors(true)

	j := job.Job{
		BaseName:  "job_err",
		Batch:     "batch_err",
		Status:    job.StatusError,
		ErrorFile: "C:/temp/error/job_err.log",
		Signature: "SIG-ERR-1",
		Title:     "Failed Record",
	}

	ShowDetail(j, tv)
	text := tv.GetText(true)

	if !strings.Contains(text, "Status:    error") {
		t.Errorf("expected Status in output, got:\n%s", text)
	}
	if !strings.Contains(text, "Error File:      C:/temp/error/job_err.log") {
		t.Errorf("expected Error File in output, got:\n%s", text)
	}
}
