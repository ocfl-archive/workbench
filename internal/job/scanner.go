package job

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

// GetBatches scans the provided folder and returns names of all direct subdirectories.
func GetBatches(folder string) ([]string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("error reading directory '%s': %w", folder, err)
	}

	var batches []string
	for _, entry := range entries {
		if entry.IsDir() {
			batches = append(batches, entry.Name())
		}
	}
	return batches, nil
}

// GetSignatures scans the batch's incoming folder for jobs, parsing metadata (info.json)
// and locating associated OCFL containers and reports.
func GetSignatures(batchName, inputF, ocflF, reportF string, logger zerolog.Logger) ([]Job, error) {
	entries, err := os.ReadDir(inputF)
	if err != nil {
		return nil, fmt.Errorf("error reading input directory '%s': %w", inputF, err)
	}

	var jobs []Job
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		baseName := entry.Name()
		folderPath := filepath.Join(inputF, baseName)

		// 1. Check for info.json inside the folder or <foldername>.json next to it
		infoPath := filepath.Join(folderPath, "info.json")
		if _, err := os.Stat(infoPath); os.IsNotExist(err) {
			infoPath = filepath.Join(folderPath, baseName+".json")
			if _, err := os.Stat(infoPath); os.IsNotExist(err) {
				logger.Warn().Str("folder", folderPath).Msg("Skipping folder: no info.json found")
				continue
			}
		}

		dataBytes, err := os.ReadFile(infoPath)
		if err != nil {
			logger.Error().Err(err).Str("file", infoPath).Msg("Error reading info file")
			continue
		}

		var meta Info
		if err := json.Unmarshal(dataBytes, &meta); err != nil {
			logger.Error().Err(err).Str("file", infoPath).Msg("Error parsing info JSON")
			continue
		}

		// 2. Check for data and metadata subfolders
		dataPath := filepath.Join(folderPath, "data")
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			dataPath = ""
		}

		metadataPath := filepath.Join(folderPath, "metadata")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			metadataPath = ""
		}

		// 3. Check for existing OCFL container (.zip)
		var ocflFile string
		cleanBaseName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		ocflZip := filepath.Join(ocflF, cleanBaseName+".zip")
		if _, err := os.Stat(ocflZip); err == nil {
			ocflFile = ocflZip
		}

		// 4. Check for existing report (.pdf)
		var reportFile string
		reportPdf := filepath.Join(reportF, cleanBaseName+".pdf")
		if _, err := os.Stat(reportPdf); err == nil {
			reportFile = reportPdf
		}

		jobs = append(jobs, Job{
			BaseName:       baseName,
			Batch:          batchName,
			InfoFile:       infoPath,
			DataFolder:     dataPath,
			MetadataFolder: metadataPath,
			OcflFile:       ocflFile,
			ReportFile:     reportFile,
			Signature:      meta.Signature,
			Title:          meta.Title,
		})
	}

	return jobs, nil
}
