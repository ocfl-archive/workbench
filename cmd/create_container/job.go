package main

import (
	"encoding/json"
	"os"
	"path"
	"slices"

	"emperror.dev/errors"
	"github.com/rs/zerolog"
)

// Job represents a container packaging task along with the status of its associated input files and generated artifacts.
type Job struct {
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
	entries, err := os.ReadDir(inputF)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", inputF)
	}
	var result = make([]Job, 0)
	var folders = make([]string, 0)
	var jsons = make([]string, 0)

	// Classify directory entries into subdirectories and JSON metadata files
	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
			continue
		}
		if path.Ext(entry.Name()) == ".json" {
			jsons = append(jsons, entry.Name())
			continue
		}
	}

	// Process each JSON file and pair it with its corresponding folder
	for _, jsonfile := range jsons {
		base := jsonfile[0 : len(jsonfile)-5]
		if !slices.Contains(folders, base) {
			logger.Warn().Msgf("No folder for %s - skipping", path.Join(inputF, jsonfile))
			continue
		}
		jsonBytes, err := os.ReadFile(path.Join(inputF, jsonfile))
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to read %s", path.Join(inputF, jsonfile))
			continue
		}
		var i = &info{}
		if err := json.Unmarshal(jsonBytes, &i); err != nil {
			logger.Error().Err(err).Msgf("Failed to unmarshal %s", path.Join(inputF, jsonfile))
			continue
		}

		job := Job{
			signature: i.Signature,
			title:     i.Title,
			infoFile:  path.Join(inputF, jsonfile),
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
