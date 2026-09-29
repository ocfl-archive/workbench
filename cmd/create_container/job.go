package main

import (
	"encoding/json"
	"os"
	"path"

	"emperror.dev/errors"
	"github.com/rs/zerolog"
)

// Job represents a container packaging task along with the status of its associated input files and generated artifacts.
type Job struct {
	baseName       string // Base identifier / folder name of the job.
	infoFile       string // Path to the JSON file containing signature and title metadata.
	dataFolder     string // Path to the folder containing data payload files, if present.
	metadataFolder string // Path to the folder containing metadata files, if present.
	ocflFile       string // Path to the generated OCFL zip archive, if it exists.
	reportFile     string // Path to the generated PDF report, if it exists.
	signature      string // Unique identifier/signature for the job.
	title          string // Human-readable title for the job.
}

// info represents the JSON metadata structure defining a job's signature and title.
type info struct {
	Signature string `json:"signature"`
	Title     string `json:"title"`
}

// getBatches scans the given root directory and returns a slice of subdirectory names representing individual batches.
func getBatches(folder string) ([]string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", folder)
	}
	var result = make([]string, 0)
	for _, entry := range entries {
		// Only directories qualify as batch folders
		if !entry.IsDir() {
			continue
		}
		result = append(result, entry.Name())
	}
	return result, nil
}

// getSignatures scans a batch directory for matching metadata JSON files and folders,
// parses job attributes, and checks for existing artifacts (data, metadata, OCFL zip, PDF report).
func getSignatures(inputF, ocflF, reportF string, logger zerolog.Logger) ([]Job, error) {
	logger.Debug().Msgf("Scanning directory %s for signatures", inputF)
	entries, err := os.ReadDir(inputF)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", inputF)
	}
	var result = make([]Job, 0)
	var folders = make([]string, 0)

	// Collect all subdirectories representing potential job folders
	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		}
	}

	// For each job folder, resolve its metadata JSON file (inside folder or sibling)
	for _, base := range folders {
		var infoFilePath string
		folderInfoPath := path.Join(inputF, base, "info.json")
		siblingInfoPath := path.Join(inputF, base+".json")

		if _, err := os.Stat(folderInfoPath); err == nil {
			infoFilePath = folderInfoPath
		} else if _, err := os.Stat(siblingInfoPath); err == nil {
			infoFilePath = siblingInfoPath
		} else {
			logger.Warn().Msgf("No metadata JSON found for folder '%s' - skipping in '%s' and '%s'", path.Join(inputF, base), folderInfoPath, siblingInfoPath)
			continue
		}

		jsonBytes, err := os.ReadFile(infoFilePath)
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to read %s", infoFilePath)
			continue
		}
		var i = &info{}
		if err := json.Unmarshal(jsonBytes, &i); err != nil {
			logger.Error().Err(err).Msgf("Failed to unmarshal %s", infoFilePath)
			continue
		}

		job := Job{
			baseName:  base,
			signature: i.Signature,
			title:     i.Title,
			infoFile:  infoFilePath,
		}

		// Check for the existence of data and metadata subfolders
		if _, err := os.Stat(path.Join(inputF, base, "data")); err == nil {
			job.dataFolder = path.Join(inputF, base, "data")
		}
		if _, err := os.Stat(path.Join(inputF, base, "metadata")); err == nil {
			job.metadataFolder = path.Join(inputF, base, "metadata")
		}

		// Check if the report PDF has already been generated
		if _, err := os.Stat(path.Join(reportF, base+".pdf")); err == nil {
			job.reportFile = path.Join(reportF, base+".pdf")
		}

		// Check if the OCFL zip container has already been generated
		if _, err := os.Stat(path.Join(ocflF, base+".zip")); err == nil {
			job.ocflFile = path.Join(ocflF, base+".zip")
		}

		result = append(result, job)
	}
	return result, nil
}
