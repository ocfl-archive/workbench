package runner

import (
	"fmt"
	"os/exec"
	"path/filepath"

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
