package web

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/ocfl-archive/workbench/internal/job"
	"github.com/ocfl-archive/workbench/internal/runner"
	"github.com/ocfl-archive/workbench/internal/util"
	"github.com/rs/zerolog"
)

// ExecutionManager manages asynchronous command executions and batch processing.
type ExecutionManager struct {
	conf        *config.WBConfig
	broadcaster *Broadcaster
	logger      zerolog.Logger
	mu          sync.Mutex
	isRunning   bool
	activeTask  string
}

// NewExecutionManager creates a new instance of ExecutionManager.
func NewExecutionManager(conf *config.WBConfig, broadcaster *Broadcaster, logger zerolog.Logger) *ExecutionManager {
	return &ExecutionManager{
		conf:        conf,
		broadcaster: broadcaster,
		logger:      logger,
	}
}

// IsRunning returns whether an execution is currently in progress.
func (em *ExecutionManager) IsRunning() (bool, string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	return em.isRunning, em.activeTask
}

// RunCommandStream executes an external command, streaming each output line to the broadcaster.
func (em *ExecutionManager) RunCommandStream(cmd *exec.Cmd, title string) (int, string) {
	em.broadcaster.AppendLog("command", fmt.Sprintf("$ %s", strings.Join(cmd.Args, " ")))
	em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-secondary font-mono text-xs">$ %s</div>`, escapeHTML(strings.Join(cmd.Args, " "))))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		msg := fmt.Sprintf("Error creating stdout pipe: %v", err)
		em.broadcaster.AppendLog("error", msg)
		em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">%s</div>`, escapeHTML(msg)))
		return -1, msg
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		msg := fmt.Sprintf("Error creating stderr pipe: %v", err)
		em.broadcaster.AppendLog("error", msg)
		em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">%s</div>`, escapeHTML(msg)))
		return -1, msg
	}

	var output strings.Builder
	if err := cmd.Start(); err != nil {
		errMsg := fmt.Sprintf("Error starting process: %v", err)
		em.broadcaster.AppendLog("error", errMsg)
		em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">%s</div>`, escapeHTML(errMsg)))
		return -1, errMsg
	}

	multiReader := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(multiReader)
	for scanner.Scan() {
		line := scanner.Text()
		output.WriteString(line)
		output.WriteString("\n")
		em.broadcaster.AppendLog("process", line)
		em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-base-content/90 font-mono text-xs leading-relaxed">%s</div>`, escapeHTML(line)))
	}

	var exitCode int
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	} else {
		exitCode = 0
	}

	return exitCode, output.String()
}

// ExecuteJobAction executes a single job lifecycle action asynchronously.
func (em *ExecutionManager) ExecuteJobAction(j job.Job, action job.ActionType) error {
	em.mu.Lock()
	if em.isRunning {
		em.mu.Unlock()
		return fmt.Errorf("eine andere Aufgabe läuft bereits (%s)", em.activeTask)
	}
	em.isRunning = true
	em.activeTask = fmt.Sprintf("%s: %s", action.Label(), j.Signature)
	em.mu.Unlock()

	go func() {
		defer func() {
			em.mu.Lock()
			em.isRunning = false
			em.activeTask = ""
			em.mu.Unlock()
			em.broadcaster.Broadcast("terminal-done", "")
			em.broadcaster.Broadcast("refresh", "")
		}()

		em.broadcaster.Broadcast("terminal-start", fmt.Sprintf("Aktion '%s' für %s gestartet", action.Label(), j.Signature))

		switch action {
		case job.ActionReset:
			em.broadcaster.AppendLog("info", fmt.Sprintf("Setze Job %s zurück...", j.Signature))
			if err := job.ResetJob(em.conf, j); err != nil {
				em.logger.Error().Err(err).Msgf("Failed to reset job %s", j.Signature)
				em.broadcaster.AppendLog("error", fmt.Sprintf("Fehler beim Zurücksetzen: %v", err))
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">Fehler beim Zurücksetzen: %v</div>`, err))
			} else {
				em.logger.Info().Msgf("Job %s successfully reset", j.Signature)
				em.broadcaster.AppendLog("info", fmt.Sprintf("Job %s erfolgreich zurückgesetzt.", j.Signature))
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-success font-mono text-xs">Job %s erfolgreich auf Incoming zurückgesetzt.</div>`, escapeHTML(j.Signature)))
			}

		case job.ActionCheckStatus:
			cmd, err := runner.StatusOCFLCmd(em.conf, j.IngestID)
			if err != nil {
				em.broadcaster.AppendLog("error", fmt.Sprintf("Statusabfrage Vorbereitung fehlgeschlagen: %v", err))
				return
			}
			exitCode, _ := em.RunCommandStream(cmd, fmt.Sprintf("Status: %s", j.Signature))
			if exitCode != 0 {
				em.logger.Error().Msgf("Failed to query status for %s (exit code %d)", j.Signature, exitCode)
			} else {
				em.logger.Info().Msgf("Status query for %s completed", j.Signature)
			}

		case job.ActionCreateOCFL:
			cmd, err := runner.CreateOCFLCmd(em.conf, j)
			if err != nil {
				em.broadcaster.AppendLog("error", fmt.Sprintf("OCFL Vorbereitung fehlgeschlagen: %v", err))
				return
			}
			exitCode, output := em.RunCommandStream(cmd, fmt.Sprintf("OCFL Erstellung: %s", j.Signature))
			if exitCode != 0 {
				em.logger.Error().Msgf("Failed to create OCFL for %s (exit code: %d)", j.Signature, exitCode)
				if err := job.HandleJobError(em.conf, j, cmd.String(), output); err != nil {
					em.logger.Error().Err(err).Msgf("Failed to handle job error for %s", j.Signature)
				}
			} else {
				em.logger.Info().Msgf("OCFL container creation for %s completed successfully", j.Signature)
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-success font-mono text-xs">OCFL-Container für %s erfolgreich erstellt.</div>`, escapeHTML(j.Signature)))
			}

		case job.ActionCreateReport:
			port, err := util.GetFreePort()
			if err != nil {
				em.broadcaster.AppendLog("error", fmt.Sprintf("Kein freier Port gefunden: %v", err))
				return
			}
			cmd, err := runner.CreateReportCmd(em.conf, j, port)
			if err != nil {
				em.broadcaster.AppendLog("error", fmt.Sprintf("Report Vorbereitung fehlgeschlagen: %v", err))
				return
			}
			exitCode, output := em.RunCommandStream(cmd, fmt.Sprintf("Report Erstellung: %s", j.Signature))
			if exitCode != 0 {
				em.logger.Error().Msgf("Failed to create report for %s (exit code: %d)", j.Signature, exitCode)
				if err := job.HandleJobError(em.conf, j, cmd.String(), output); err != nil {
					em.logger.Error().Err(err).Msgf("Failed to handle job error for %s", j.Signature)
				}
			} else {
				em.logger.Info().Msgf("Report for %s completed successfully", j.Signature)
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-success font-mono text-xs">Report für %s erfolgreich erstellt.</div>`, escapeHTML(j.Signature)))
			}

		case job.ActionValidate:
			cmd, err := runner.ValidateOCFLCmd(em.conf, j)
			if err != nil {
				em.broadcaster.AppendLog("error", fmt.Sprintf("Validierung Vorbereitung fehlgeschlagen: %v", err))
				return
			}
			exitCode, output := em.RunCommandStream(cmd, fmt.Sprintf("OCFL Validierung: %s", j.Signature))
			if exitCode != 0 {
				em.logger.Error().Msgf("Failed to validate OCFL for %s (exit code: %d)", j.Signature, exitCode)
				if err := job.HandleJobError(em.conf, j, cmd.String(), output); err != nil {
					em.logger.Error().Err(err).Msgf("Failed to handle job error for %s", j.Signature)
				}
			} else {
				em.logger.Info().Msgf("Validation for %s completed successfully", j.Signature)
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-success font-mono text-xs">OCFL-Container für %s erfolgreich validiert.</div>`, escapeHTML(j.Signature)))
			}

		case job.ActionIngest:
			cmd, resultJsonPath, err := runner.IngestOCFLCmd(em.conf, j)
			if err != nil {
				em.broadcaster.AppendLog("error", fmt.Sprintf("Ingest Vorbereitung fehlgeschlagen: %v", err))
				return
			}
			exitCode, output := em.RunCommandStream(cmd, fmt.Sprintf("OCFL Ingest: %s", j.Signature))
			hasError := exitCode != 0
			var errMsg string
			if !hasError {
				var checkErr error
				hasError, errMsg, checkErr = job.CheckIngestResult(resultJsonPath)
				if checkErr != nil {
					em.logger.Warn().Err(checkErr).Msgf("Could not evaluate result JSON for %s", j.Signature)
				}
			}

			if hasError {
				if errMsg != "" {
					em.logger.Error().Msgf("Failed to ingest OCFL container for %s (error: %s)", j.Signature, errMsg)
					em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">Ingest-Fehler: %s</div>`, escapeHTML(errMsg)))
				} else {
					em.logger.Error().Msgf("Failed to ingest OCFL container for %s (exit code: %d)", j.Signature, exitCode)
				}
				if err := job.HandleIngestError(em.conf, j, cmd.String(), output, resultJsonPath); err != nil {
					em.logger.Error().Err(err).Msgf("Failed to handle ingest error for %s", j.Signature)
				}
			} else {
				em.logger.Info().Msgf("OCFL ingest for %s completed successfully", j.Signature)
				if err := job.ArchiveJob(em.conf, j, resultJsonPath); err != nil {
					em.logger.Error().Err(err).Msgf("Failed to archive job for %s", j.Signature)
				} else {
					em.logger.Info().Msgf("Successfully archived job artifacts for %s", j.Signature)
					em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-success font-mono text-xs">Job %s erfolgreich archiviert.</div>`, escapeHTML(j.Signature)))
				}
			}
		}
	}()

	return nil
}

// ExecuteBatch runs the full preservation pipeline for all eligible incoming jobs in a batch.
func (em *ExecutionManager) ExecuteBatch(batchName string) error {
	em.mu.Lock()
	if em.isRunning {
		em.mu.Unlock()
		return fmt.Errorf("eine andere Aufgabe läuft bereits (%s)", em.activeTask)
	}
	em.isRunning = true
	em.activeTask = fmt.Sprintf("Batch: %s", batchName)
	em.mu.Unlock()

	go func() {
		defer func() {
			em.mu.Lock()
			em.isRunning = false
			em.activeTask = ""
			em.mu.Unlock()
			em.broadcaster.Broadcast("terminal-done", "")
			em.broadcaster.Broadcast("refresh", "")
		}()

		em.broadcaster.Broadcast("terminal-start", fmt.Sprintf("Batch-Verarbeitung für '%s' gestartet", batchName))
		em.broadcaster.AppendLog("info", fmt.Sprintf("Starte Batch-Pipeline für '%s'...", batchName))

		inputFolder := em.conf.GetInputFolder(batchName)
		ocflFolder := em.conf.GetOcflFolder(batchName)
		reportFolder := em.conf.GetReportFolder(batchName)
		errorFolder := em.conf.GetErrorFolder(batchName)
		archivedFolder := em.conf.GetArchivedFolder(batchName)

		jobs, err := job.GetSignatures(batchName, inputFolder, ocflFolder, reportFolder, errorFolder, archivedFolder, em.logger)
		if err != nil {
			em.logger.Error().Err(err).Msgf("Failed to read jobs for batch %s", batchName)
			em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">Fehler beim Lesen der Jobs: %v</div>`, err))
			return
		}

		var eligibleJobs []job.Job
		for _, j := range jobs {
			if j.Signature != "" && j.Status == job.StatusIncoming {
				eligibleJobs = append(eligibleJobs, j)
			}
		}

		if len(eligibleJobs) == 0 {
			em.broadcaster.Broadcast("terminal-line", `<div class="text-warning font-mono text-xs">Keine ausstehenden Jobs im Batch gefunden (alle bereits archiviert oder fehlerhaft).</div>`)
			return
		}

		em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-info font-mono text-xs font-bold">Verarbeite %d Jobs in Batch '%s'...</div>`, len(eligibleJobs), escapeHTML(batchName)))

		var successCount, errorCount int
		for i, currentJob := range eligibleJobs {
			idx := i + 1
			em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="divider my-2 text-xs font-bold opacity-75">[%d/%d] Job: %s (%s)</div>`, idx, len(eligibleJobs), escapeHTML(currentJob.Signature), escapeHTML(currentJob.BaseName)))

			// 1. Create OCFL if missing
			if currentJob.OcflFile == "" {
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-info font-mono text-xs">1. Erstelle OCFL-Container für %s...</div>`, escapeHTML(currentJob.Signature)))
				cmd, err := runner.CreateOCFLCmd(em.conf, currentJob)
				if err != nil {
					em.logger.Error().Err(err).Msgf("Failed to prepare OCFL creation for %s", currentJob.Signature)
					em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">Fehler bei Vorbereitung OCFL: %v</div>`, err))
					errorCount++
					continue
				}

				exitCode, output := em.RunCommandStream(cmd, "OCFL")
				if exitCode != 0 {
					em.logger.Error().Msgf("Failed to create OCFL for %s", currentJob.Signature)
					if err := job.HandleJobError(em.conf, currentJob, cmd.String(), output); err != nil {
						em.logger.Error().Err(err).Msgf("Failed to handle job error for %s", currentJob.Signature)
					}
					errorCount++
					continue
				}
				currentJob.OcflFile = currentJob.ExpectedZipPath(em.conf)
			}

			// 2. Validate OCFL
			em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-info font-mono text-xs">2. Validiere OCFL-Container für %s...</div>`, escapeHTML(currentJob.Signature)))
			valCmd, err := runner.ValidateOCFLCmd(em.conf, currentJob)
			if err != nil {
				em.logger.Error().Err(err).Msgf("Failed to prepare OCFL validation for %s", currentJob.Signature)
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">Fehler bei Vorbereitung Validierung: %v</div>`, err))
				errorCount++
				continue
			}

			valExitCode, valOutput := em.RunCommandStream(valCmd, "Validate")
			if valExitCode != 0 {
				em.logger.Error().Msgf("Failed to validate OCFL for %s", currentJob.Signature)
				if err := job.HandleJobError(em.conf, currentJob, valCmd.String(), valOutput); err != nil {
					em.logger.Error().Err(err).Msgf("Failed to handle job error for %s", currentJob.Signature)
				}
				errorCount++
				continue
			}

			// 3. Create Report if missing
			if currentJob.ReportFile == "" {
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-info font-mono text-xs">3. Erstelle Report für %s...</div>`, escapeHTML(currentJob.Signature)))
				port, err := util.GetFreePort()
				if err != nil {
					em.logger.Error().Err(err).Msgf("Failed to get free port for %s", currentJob.Signature)
					errorCount++
					continue
				}
				repCmd, err := runner.CreateReportCmd(em.conf, currentJob, port)
				if err != nil {
					em.logger.Error().Err(err).Msgf("Failed to prepare report for %s", currentJob.Signature)
					errorCount++
					continue
				}

				repExitCode, repOutput := em.RunCommandStream(repCmd, "Report")
				if repExitCode != 0 {
					em.logger.Error().Msgf("Failed to create report for %s", currentJob.Signature)
					if err := job.HandleJobError(em.conf, currentJob, repCmd.String(), repOutput); err != nil {
						em.logger.Error().Err(err).Msgf("Failed to handle job error for %s", currentJob.Signature)
					}
					errorCount++
					continue
				}
				currentJob.ReportFile = currentJob.ExpectedReportPath(em.conf)
			}

			// 4. Ingest OCFL
			em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-info font-mono text-xs">4. Starte Ingest für %s...</div>`, escapeHTML(currentJob.Signature)))
			ingestCmd, resultJsonPath, err := runner.IngestOCFLCmd(em.conf, currentJob)
			if err != nil {
				em.logger.Error().Err(err).Msgf("Failed to prepare ingest for %s", currentJob.Signature)
				errorCount++
				continue
			}

			ingExitCode, ingOutput := em.RunCommandStream(ingestCmd, "Ingest")
			hasError := ingExitCode != 0
			var errMsg string
			if !hasError {
				var checkErr error
				hasError, errMsg, checkErr = job.CheckIngestResult(resultJsonPath)
				if checkErr != nil {
					em.logger.Warn().Err(checkErr).Msgf("Could not evaluate result JSON for %s", currentJob.Signature)
				}
			}

			if hasError {
				if errMsg != "" {
					em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-error font-mono text-xs">Ingest-Fehler: %s</div>`, escapeHTML(errMsg)))
				}
				if err := job.HandleIngestError(em.conf, currentJob, ingestCmd.String(), ingOutput, resultJsonPath); err != nil {
					em.logger.Error().Err(err).Msgf("Failed to handle ingest error for %s", currentJob.Signature)
				}
				errorCount++
				continue
			}

			if err := job.ArchiveJob(em.conf, currentJob, resultJsonPath); err != nil {
				em.logger.Error().Err(err).Msgf("Failed to archive job for %s", currentJob.Signature)
				errorCount++
			} else {
				em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(`<div class="text-success font-mono text-xs">Job %s erfolgreich archiviert.</div>`, escapeHTML(currentJob.Signature)))
				successCount++
			}
			time.Sleep(100 * time.Millisecond)
		}

		em.broadcaster.Broadcast("terminal-line", fmt.Sprintf(
			`<div class="mt-4 p-3 rounded-lg bg-base-200 text-xs font-bold">Batch-Verarbeitung abgeschlossen: <span class="text-success">%d erfolgreich</span>, <span class="text-error">%d fehlerhaft</span></div>`,
			successCount, errorCount,
		))
	}()

	return nil
}
