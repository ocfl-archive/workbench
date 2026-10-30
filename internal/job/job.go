// Package job encapsulates the domain model, directory scanner, path resolver,
// and lifecycle management for digitization and preservation jobs.
package job

import (
	"path/filepath"

	"github.com/ocfl-archive/workbench/internal/config"
)

// JobStatus represents the processing status and location of a job within a batch.
type JobStatus string

const (
	// StatusIncoming indicates a job is in the incoming folder awaiting processing or ingest.
	StatusIncoming JobStatus = "incoming"
	// StatusError indicates a job failed during processing or ingest and was moved to the error folder.
	StatusError JobStatus = "error"
	// StatusArchived indicates a job has been successfully ingested and archived.
	StatusArchived JobStatus = "archived"
)

// ActionType represents an executable action on a job.
type ActionType string

const (
	// ActionCreateOCFL triggers OCFL container zip creation.
	ActionCreateOCFL ActionType = "create_ocfl"
	// ActionCreateReport triggers PDF validation report generation.
	ActionCreateReport ActionType = "create_report"
	// ActionValidate triggers OCFL package compliance validation.
	ActionValidate ActionType = "validate"
	// ActionIngest triggers package transmission to the preservation repository via ONA.
	ActionIngest ActionType = "ingest"
	// ActionCheckStatus triggers polling the ingest status from ONA using the IngestID.
	ActionCheckStatus ActionType = "check_status"
)

// Label returns the human-readable UI button label for the action.
func (a ActionType) Label() string {
	switch a {
	case ActionCreateOCFL:
		return "OCFL erstellen"
	case ActionCreateReport:
		return "Report erstellen"
	case ActionValidate:
		return "Validate"
	case ActionIngest:
		return "Ingest"
	case ActionCheckStatus:
		return "Status abfragen"
	default:
		return string(a)
	}
}

// Job represents a single digitization/preservation unit within a batch.
type Job struct {
	// BaseName is the folder name or identifier of the job item.
	BaseName string
	// Batch is the name of the batch directory containing this job.
	Batch string
	// Status is the current lifecycle status (incoming, error, archived).
	Status JobStatus
	// IncomingFolder is the path to the incoming source directory for this job.
	IncomingFolder string
	// InfoFile is the path to the info.json descriptive metadata file.
	InfoFile string
	// DataFolder is the path to the payload data subfolder (if present).
	DataFolder string
	// MetadataFolder is the path to the metadata subfolder (if present).
	MetadataFolder string
	// OcflFile is the path to the generated OCFL .zip container.
	OcflFile string
	// ChecksumFile is the path to the generated OCFL .sha512 checksum file.
	ChecksumFile string
	// ReportFile is the path to the generated validation .pdf report.
	ReportFile string
	// UploadFile is the path to the .upload.json marker file.
	UploadFile string
	// ErrorFile is the path to the error .log file.
	ErrorFile string
	// IngestID is the unique archival ingest identifier returned by ONA.
	IngestID string
	// Signature is the archival call number / signature from metadata.
	Signature string
	// Title is the descriptive title from metadata.
	Title string
}

// AvailableActions returns the list of actions permitted for the job given its current state.
func (j Job) AvailableActions() []ActionType {
	if j.Signature == "" || j.Status == StatusError {
		return nil
	}
	if j.Status == StatusArchived {
		if j.IngestID != "" {
			return []ActionType{ActionCheckStatus}
		}
		return nil
	}
	if j.Status == StatusIncoming {
		if j.OcflFile == "" {
			return []ActionType{ActionCreateOCFL}
		}
		var actions []ActionType
		if j.ReportFile == "" {
			actions = append(actions, ActionCreateReport)
		}
		actions = append(actions, ActionValidate, ActionIngest)
		return actions
	}
	return nil
}

// ExpectedZipName returns the standard OCFL zip filename for the job.
func (j Job) ExpectedZipName() string {
	return j.BaseName + ".zip"
}

// ExpectedChecksumName returns the standard OCFL sha512 checksum filename for the job.
func (j Job) ExpectedChecksumName() string {
	return j.ExpectedZipName() + ".sha512"
}

// ExpectedReportName returns the standard validation PDF report filename for the job.
func (j Job) ExpectedReportName() string {
	return j.BaseName + ".pdf"
}

// ExpectedResultJsonName returns the standard ONA ingest result JSON filename for the job.
func (j Job) ExpectedResultJsonName() string {
	return j.BaseName + "_ona.json"
}

// ExpectedUploadJsonName returns the standard upload marker JSON filename for the job.
func (j Job) ExpectedUploadJsonName() string {
	return j.BaseName + ".upload.json"
}

// ExpectedErrorLogName returns the standard error log filename for the job.
func (j Job) ExpectedErrorLogName() string {
	return j.BaseName + ".log"
}

// ExpectedZipPath returns the full path where the OCFL zip archive is stored.
func (j Job) ExpectedZipPath(conf *config.WBConfig) string {
	if conf == nil {
		return ""
	}
	return filepath.Join(conf.GetOcflFolder(j.Batch), j.ExpectedZipName())
}

// ExpectedChecksumPath returns the full path where the OCFL sha512 checksum file is stored.
func (j Job) ExpectedChecksumPath(conf *config.WBConfig) string {
	if conf == nil {
		return ""
	}
	return filepath.Join(conf.GetOcflFolder(j.Batch), j.ExpectedChecksumName())
}

// ExpectedReportPath returns the full path where the validation PDF report is stored.
func (j Job) ExpectedReportPath(conf *config.WBConfig) string {
	if conf == nil {
		return ""
	}
	return filepath.Join(conf.GetReportFolder(j.Batch), j.ExpectedReportName())
}

// ExpectedResultJsonPath returns the full path where the ONA ingest result JSON is stored.
func (j Job) ExpectedResultJsonPath(conf *config.WBConfig) string {
	if conf == nil {
		return ""
	}
	return filepath.Join(conf.GetOcflFolder(j.Batch), j.ExpectedResultJsonName())
}

// ExpectedUploadPath returns the full path where the upload marker JSON is stored.
func (j Job) ExpectedUploadPath(conf *config.WBConfig) string {
	if conf == nil {
		return ""
	}
	return filepath.Join(conf.GetOcflFolder(j.Batch), j.ExpectedUploadJsonName())
}

// ExpectedErrorLogPath returns the full path where the job error log is stored.
func (j Job) ExpectedErrorLogPath(conf *config.WBConfig) string {
	if conf == nil {
		return ""
	}
	return filepath.Join(conf.GetErrorFolder(j.Batch), j.ExpectedErrorLogName())
}

// Info represents the descriptive metadata decoded from an info.json file.
type Info struct {
	Signature string `json:"signature"`
	Title     string `json:"title"`
}

// UploadInfo represents the upload metadata stored in .upload.json.
type UploadInfo struct {
	Date string `json:"date"`
}
