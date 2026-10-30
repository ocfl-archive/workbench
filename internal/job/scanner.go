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

// ExtractIngestID extracts the "id" field from an ingest result JSON byte slice.
// It handles string, float64, and integer representations and returns an empty string if missing or invalid.
func ExtractIngestID(data []byte) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	val, ok := payload["id"]
	if !ok || val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// scanDirectory scans a specific batch subfolder (incoming, error, archived) for jobs.
func scanDirectory(batchName, folder, ocflF, reportF string, status JobStatus, logger zerolog.Logger) ([]Job, error) {
	if folder == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading directory '%s': %w", folder, err)
	}

	var jobs []Job
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		baseName := entry.Name()
		folderPath := filepath.Join(folder, baseName)

		// 1. Check for info.json inside the folder or <foldername>.json inside / next to it
		infoPath := filepath.Join(folderPath, "info.json")
		if _, err := os.Stat(infoPath); os.IsNotExist(err) {
			infoPath = filepath.Join(folderPath, baseName+".json")
			if _, err := os.Stat(infoPath); os.IsNotExist(err) {
				infoPath = filepath.Join(folder, baseName+".json")
				if _, err := os.Stat(infoPath); os.IsNotExist(err) {
					logger.Warn().Str("folder", folderPath).Msg("Skipping folder: no info.json found")
					continue
				}
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

		cleanBaseName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		// 3. Check for existing OCFL container (.zip) and checksum (.sha512)
		var ocflFile string
		var checksumFile string
		switch status {
		case StatusArchived:
			ocflZip := filepath.Join(folder, cleanBaseName+".zip")
			if _, err := os.Stat(ocflZip); err == nil {
				ocflFile = ocflZip
			} else {
				ocflZip = filepath.Join(ocflF, cleanBaseName+".zip")
				if _, err := os.Stat(ocflZip); err == nil {
					ocflFile = ocflZip
				}
			}

			candidates := []string{
				filepath.Join(folder, cleanBaseName+".zip.sha512"),
				filepath.Join(folder, cleanBaseName+".sha512"),
				filepath.Join(ocflF, cleanBaseName+".zip.sha512"),
				filepath.Join(ocflF, cleanBaseName+".sha512"),
			}
			for _, candidate := range candidates {
				if _, err := os.Stat(candidate); err == nil {
					checksumFile = candidate
					break
				}
			}
		default:
			ocflZip := filepath.Join(ocflF, cleanBaseName+".zip")
			if _, err := os.Stat(ocflZip); err == nil {
				ocflFile = ocflZip
			}

			candidates := []string{
				filepath.Join(ocflF, cleanBaseName+".zip.sha512"),
				filepath.Join(ocflF, cleanBaseName+".sha512"),
			}
			for _, candidate := range candidates {
				if _, err := os.Stat(candidate); err == nil {
					checksumFile = candidate
					break
				}
			}
		}

		// 4. Check for existing report (.pdf)
		var reportFile string
		reportPdf := filepath.Join(reportF, cleanBaseName+".pdf")
		if _, err := os.Stat(reportPdf); err == nil {
			reportFile = reportPdf
		}

		// 5. Check for upload marker, error log, or ingest ID
		var uploadFile string
		var errorFile string
		var ingestID string

		switch status {
		case StatusIncoming:
			uploadJson := filepath.Join(ocflF, cleanBaseName+".upload.json")
			if _, err := os.Stat(uploadJson); err == nil {
				uploadFile = uploadJson
			} else {
				ingestJson := filepath.Join(ocflF, cleanBaseName+"_ona.json")
				if _, err := os.Stat(ingestJson); err == nil {
					uploadFile = ingestJson
				} else {
					legacyJson := filepath.Join(ocflF, cleanBaseName+".json")
					if _, err := os.Stat(legacyJson); err == nil {
						uploadFile = legacyJson
					}
				}
			}
		case StatusError:
			errLog := filepath.Join(folder, cleanBaseName+".log")
			if _, err := os.Stat(errLog); err == nil {
				errorFile = errLog
			} else {
				errLog = filepath.Join(folder, baseName+".log")
				if _, err := os.Stat(errLog); err == nil {
					errorFile = errLog
				}
			}
		case StatusArchived:
			archivedJson := filepath.Join(folder, cleanBaseName+"_ona.json")
			if data, err := os.ReadFile(archivedJson); err == nil {
				ingestID = ExtractIngestID(data)
			} else {
				archivedJson = filepath.Join(folder, baseName+"_ona.json")
				if data, err := os.ReadFile(archivedJson); err == nil {
					ingestID = ExtractIngestID(data)
				}
			}
		}

		jobs = append(jobs, Job{
			BaseName:       baseName,
			Batch:          batchName,
			Status:         status,
			IncomingFolder: folderPath,
			InfoFile:       infoPath,
			DataFolder:     dataPath,
			MetadataFolder: metadataPath,
			OcflFile:       ocflFile,
			ChecksumFile:   checksumFile,
			ReportFile:     reportFile,
			UploadFile:     uploadFile,
			ErrorFile:      errorFile,
			IngestID:       ingestID,
			Signature:      meta.Signature,
			Title:          meta.Title,
		})
	}

	return jobs, nil
}

// GetSignatures scans the batch's incoming, error, and archived folders for jobs, parsing metadata (info.json)
// and locating associated OCFL containers, reports, error logs, and ingest IDs.
func GetSignatures(batchName, inputF, ocflF, reportF, errorF, archivedF string, logger zerolog.Logger) ([]Job, error) {
	var allJobs []Job

	incomingJobs, err := scanDirectory(batchName, inputF, ocflF, reportF, StatusIncoming, logger)
	if err != nil {
		return nil, err
	}
	allJobs = append(allJobs, incomingJobs...)

	errorJobs, err := scanDirectory(batchName, errorF, ocflF, reportF, StatusError, logger)
	if err != nil {
		return nil, err
	}
	allJobs = append(allJobs, errorJobs...)

	archivedJobs, err := scanDirectory(batchName, archivedF, ocflF, reportF, StatusArchived, logger)
	if err != nil {
		return nil, err
	}
	allJobs = append(allJobs, archivedJobs...)

	return allJobs, nil
}
