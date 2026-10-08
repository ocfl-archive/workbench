package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/ocfl-archive/workbench/internal/job"
	"github.com/ocfl-archive/workbench/internal/util"
)

// CreateOCFLCmd builds the exec.Cmd for creating an OCFL container using gocfl.
func CreateOCFLCmd(conf *config.WBConfig, j job.Job) (*exec.Cmd, error) {
	ocflDir := conf.GetOcflFolder(j.Batch)
	if err := util.EnsureTargetDirectory(ocflDir); err != nil {
		return nil, fmt.Errorf("failed to ensure OCFL target directory '%s': %w", ocflDir, err)
	}

	zipName := fmt.Sprintf("%s.zip", j.BaseName)
	cmd := exec.Command("gocfl",
		"create",
		filepath.Join(ocflDir, zipName),
		j.DataFolder,
		fmt.Sprintf("metadata:%s", j.MetadataFolder),
		"-i", j.Signature,
		"--ext-NNNN-metafile-source", j.InfoFile,
	)
	return cmd, nil
}

// CreateReportCmd builds the exec.Cmd for generating a PDF report for an OCFL container using gocfl.
func CreateReportCmd(conf *config.WBConfig, j job.Job, port int) (*exec.Cmd, error) {
	reportDir := conf.GetReportFolder(j.Batch)
	if err := util.EnsureTargetDirectory(reportDir); err != nil {
		return nil, fmt.Errorf("failed to ensure report target directory '%s': %w", reportDir, err)
	}

	ocflDir := conf.GetOcflFolder(j.Batch)
	zipName := fmt.Sprintf("%s.zip", j.BaseName)
	pdfName := fmt.Sprintf("%s.pdf", j.BaseName)

	cmd := exec.Command("gocfl",
		"display",
		filepath.Join(ocflDir, zipName),
		"--display-fullreport", filepath.Join(reportDir, pdfName),
		"--display-id", j.Signature,
		"-a", fmt.Sprintf("localhost:%d", port),
		"-e", fmt.Sprintf("http://localhost:%d", port),
	)
	return cmd, nil
}

// ValidateOCFLCmd builds the exec.Cmd for validating an OCFL container using gocfl.
func ValidateOCFLCmd(conf *config.WBConfig, j job.Job) (*exec.Cmd, error) {
	ocflDir := conf.GetOcflFolder(j.Batch)
	zipName := fmt.Sprintf("%s.zip", j.BaseName)

	cmd := exec.Command("gocfl",
		"validate",
		filepath.Join(ocflDir, zipName),
	)
	return cmd, nil
}

// IngestOCFLCmd builds the exec.Cmd for ingesting an OCFL container using ona.
func IngestOCFLCmd(conf *config.WBConfig, j job.Job) (*exec.Cmd, error) {
	ocflDir := conf.GetOcflFolder(j.Batch)
	zipName := fmt.Sprintf("%s.zip", j.BaseName)
	zipPath := filepath.Join(ocflDir, zipName)

	args := []string{"ingest", "-p", zipPath}
	if conf.Ona.Config != "" {
		args = append(args, "-c", conf.Ona.Config)
	}

	cmd := exec.Command("ona", args...)
	return cmd, nil
}

// CreateUploadRecord creates a .upload.json file for the given job containing the upload timestamp.
func CreateUploadRecord(conf *config.WBConfig, j job.Job) error {
	ocflDir := conf.GetOcflFolder(j.Batch)
	if err := util.EnsureTargetDirectory(ocflDir); err != nil {
		return fmt.Errorf("failed to ensure OCFL target directory '%s': %w", ocflDir, err)
	}

	uploadPath := filepath.Join(ocflDir, fmt.Sprintf("%s.upload.json", j.BaseName))
	uploadInfo := job.UploadInfo{
		Date: time.Now().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(uploadInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal upload info: %w", err)
	}

	if err := os.WriteFile(uploadPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write upload file '%s': %w", uploadPath, err)
	}

	return nil
}
