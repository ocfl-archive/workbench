package ui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/ocfl-archive/workbench/internal/job"
	"github.com/ocfl-archive/workbench/internal/runner"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

// queueUpdateDraw executes f on the application's event loop if app is non-nil, or immediately if app is nil.
func queueUpdateDraw(app *tview.Application, f func()) {
	if app != nil {
		app.QueueUpdateDraw(f)
	} else {
		f()
	}
}

// RunCommandStream executes a command and streams its output into modalText, returning exit code and combined output.
func RunCommandStream(app *tview.Application, modalText *tview.TextView, cmd *exec.Cmd) (int, string) {
	var output strings.Builder
	queueUpdateDraw(app, func() {
		fmt.Fprintln(tview.ANSIWriter(modalText), strings.Join(cmd.Args, " "))
		modalText.ScrollToEnd()
	})

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	var exitCode int
	if err := cmd.Start(); err != nil {
		exitCode = -1
		errMsg := fmt.Sprintf("Error starting process: %v\n", err)
		output.WriteString(errMsg)
		queueUpdateDraw(app, func() {
			fmt.Fprintf(tview.ANSIWriter(modalText), "[red]%s[white]", errMsg)
			modalText.ScrollToEnd()
		})
	} else {
		multiReader := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(multiReader)
		for scanner.Scan() {
			line := scanner.Text()
			output.WriteString(line)
			output.WriteString("\n")
			queueUpdateDraw(app, func() {
				fmt.Fprintf(tview.ANSIWriter(modalText), "%s\n", line)
				modalText.ScrollToEnd()
			})
		}
		if err := cmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		} else {
			exitCode = 0
		}
	}
	return exitCode, output.String()
}

// RunCommandInModal encapsulates modal dialog creation and process output streaming.
func RunCommandInModal(
	app *tview.Application,
	pages *tview.Pages,
	title string,
	startMsg string,
	cmd *exec.Cmd,
	onFinish func(exitCode int, output string),
) {
	modalText := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	modalText.SetBorder(true).SetTitle(title)

	modalBox := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(modalText, 0, 8, true).
			AddItem(nil, 0, 1, false), 0, 8, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("execModal", modalBox, true, true)
	if app != nil {
		app.SetFocus(modalText)
	}

	go func() {
		queueUpdateDraw(app, func() {
			fmt.Fprintf(tview.ANSIWriter(modalText), "%s\n\n", startMsg)
			modalText.ScrollToEnd()
		})

		exitCode, output := RunCommandStream(app, modalText, cmd)

		queueUpdateDraw(app, func() {
			fmt.Fprintf(tview.ANSIWriter(modalText), "\n[green]Process finished. Press [yellow]ESC[green] or [yellow]ENTER[green] to close...[white]\n")
			modalText.ScrollToEnd()

			modalText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					pages.RemovePage("execModal")
					if onFinish != nil {
						onFinish(exitCode, output)
					}
					return nil
				}
				return event
			})
		})
	}()
}

// RunBatchInModal handles sequential batch job execution (OCFL creation and ingest) in a modal dialog.
func RunBatchInModal(
	app *tview.Application,
	pages *tview.Pages,
	conf *config.WBConfig,
	batchName string,
	logger zerolog.Logger,
	onFinish func(),
) {
	modalText := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	modalText.SetBorder(true).SetTitle(fmt.Sprintf(" Batch-Verarbeitung: %s ", batchName))

	modalBox := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(modalText, 0, 8, true).
			AddItem(nil, 0, 1, false), 0, 8, true).
		AddItem(nil, 0, 1, false)

	pages.AddPage("execModal", modalBox, true, true)
	if app != nil {
		app.SetFocus(modalText)
	}

	go func() {
		inputFolder := conf.GetInputFolder(batchName)
		ocflFolder := conf.GetOcflFolder(batchName)
		reportFolder := conf.GetReportFolder(batchName)
		errorFolder := conf.GetErrorFolder(batchName)
		archivedFolder := conf.GetArchivedFolder(batchName)

		jobs, err := job.GetSignatures(batchName, inputFolder, ocflFolder, reportFolder, errorFolder, archivedFolder, logger)
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to read jobs for batch %s", batchName)
			queueUpdateDraw(app, func() {
				fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler beim Lesen der Jobs für Batch '%s': %v[white]\n", batchName, err)
				modalText.ScrollToEnd()
			})
		} else {
			var eligibleJobs []job.Job
			for _, j := range jobs {
				if j.Signature != "" && j.Status == job.StatusIncoming {
					eligibleJobs = append(eligibleJobs, j)
				}
			}

			if len(eligibleJobs) == 0 {
				queueUpdateDraw(app, func() {
					fmt.Fprintf(tview.ANSIWriter(modalText), "[yellow]Keine ausstehenden Jobs im Batch '%s' gefunden (alle Jobs sind bereits archiviert oder fehlerhaft).[white]\n", batchName)
					modalText.ScrollToEnd()
				})
			} else {
				queueUpdateDraw(app, func() {
					fmt.Fprintf(tview.ANSIWriter(modalText), "[yellow]Starte Batch-Verarbeitung für '%s' (%d Jobs gefunden)...[white]\n\n", batchName, len(eligibleJobs))
					modalText.ScrollToEnd()
				})

				var successCount, errorCount int
				for i, currentJob := range eligibleJobs {
					idx := i + 1
					queueUpdateDraw(app, func() {
						fmt.Fprintf(tview.ANSIWriter(modalText), "[yellow]=== [%d/%d] Job: %s (%s) ===[white]\n", idx, len(eligibleJobs), currentJob.Signature, currentJob.BaseName)
						modalText.ScrollToEnd()
					})

					// Step 1: Create OCFL if missing
					if currentJob.OcflFile == "" {
						queueUpdateDraw(app, func() {
							fmt.Fprintf(tview.ANSIWriter(modalText), "[yellow]Erstelle OCFL-Container für %s...[white]\n", currentJob.Signature)
							modalText.ScrollToEnd()
						})

						cmd, err := runner.CreateOCFLCmd(conf, currentJob)
						if err != nil {
							logger.Error().Err(err).Msgf("Failed to prepare OCFL creation for %s", currentJob.Signature)
							queueUpdateDraw(app, func() {
								fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler bei der Vorbereitung der OCFL-Erstellung für %s: %v[white]\n\n", currentJob.Signature, err)
								modalText.ScrollToEnd()
							})
							errorCount++
							continue
						}

						exitCode, output := RunCommandStream(app, modalText, cmd)
						if exitCode != 0 {
							logger.Error().Msgf("Failed to create OCFL container for %s (exit code: %d)", currentJob.Signature, exitCode)
							if err := job.HandleJobError(conf, currentJob, cmd.String(), output); err != nil {
								logger.Error().Err(err).Msgf("Failed to handle job error for %s", currentJob.Signature)
							}
							queueUpdateDraw(app, func() {
								fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler beim Erstellen des OCFL-Containers für %s (Exit-Code: %d)[white]\n\n", currentJob.Signature, exitCode)
								modalText.ScrollToEnd()
							})
							errorCount++
							continue
						}

						logger.Info().Msgf("OCFL container creation for %s completed successfully", currentJob.Signature)
						queueUpdateDraw(app, func() {
							fmt.Fprintf(tview.ANSIWriter(modalText), "[green]OCFL-Container für %s erfolgreich erstellt.[white]\n", currentJob.Signature)
							modalText.ScrollToEnd()
						})
						currentJob.OcflFile = currentJob.ExpectedZipPath(conf)
					}

					// Step 2: Validate OCFL
					queueUpdateDraw(app, func() {
						fmt.Fprintf(tview.ANSIWriter(modalText), "[yellow]Validiere OCFL-Container für %s...[white]\n", currentJob.Signature)
						modalText.ScrollToEnd()
					})

					valCmd, err := runner.ValidateOCFLCmd(conf, currentJob)
					if err != nil {
						logger.Error().Err(err).Msgf("Failed to prepare OCFL validation for %s", currentJob.Signature)
						queueUpdateDraw(app, func() {
							fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler bei der Vorbereitung der OCFL-Validierung für %s: %v[white]\n\n", currentJob.Signature, err)
							modalText.ScrollToEnd()
						})
						errorCount++
						continue
					}

					valExitCode, valOutput := RunCommandStream(app, modalText, valCmd)
					if valExitCode != 0 {
						logger.Error().Msgf("Failed to validate OCFL container for %s (exit code: %d)", currentJob.Signature, valExitCode)
						if err := job.HandleJobError(conf, currentJob, valCmd.String(), valOutput); err != nil {
							logger.Error().Err(err).Msgf("Failed to handle job error for %s", currentJob.Signature)
						}
						queueUpdateDraw(app, func() {
							fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler bei der OCFL-Validierung für %s (Exit-Code: %d)[white]\n\n", currentJob.Signature, valExitCode)
							modalText.ScrollToEnd()
						})
						errorCount++
						continue
					}

					logger.Info().Msgf("OCFL validation for %s completed successfully", currentJob.Signature)
					queueUpdateDraw(app, func() {
						fmt.Fprintf(tview.ANSIWriter(modalText), "[green]OCFL-Container für %s erfolgreich validiert.[white]\n", currentJob.Signature)
						modalText.ScrollToEnd()
					})

					// Step 3: Ingest OCFL
					queueUpdateDraw(app, func() {
						fmt.Fprintf(tview.ANSIWriter(modalText), "[yellow]Starte OCFL-Ingest für %s...[white]\n", currentJob.Signature)
						modalText.ScrollToEnd()
					})

					cmd, resultJsonPath, err := runner.IngestOCFLCmd(conf, currentJob)
					if err != nil {
						logger.Error().Err(err).Msgf("Failed to prepare OCFL ingest for %s", currentJob.Signature)
						queueUpdateDraw(app, func() {
							fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler bei der Vorbereitung des Ingests für %s: %v[white]\n\n", currentJob.Signature, err)
							modalText.ScrollToEnd()
						})
						errorCount++
						continue
					}

					exitCode, output := RunCommandStream(app, modalText, cmd)
					hasError := exitCode != 0
					var errMsg string
					if !hasError {
						var checkErr error
						hasError, errMsg, checkErr = job.CheckIngestResult(resultJsonPath)
						if checkErr != nil {
							logger.Warn().Err(checkErr).Msgf("Could not evaluate result JSON for %s", currentJob.Signature)
						}
					}

					if hasError {
						if errMsg != "" {
							logger.Error().Msgf("Failed to ingest OCFL container for %s (error: %s)", currentJob.Signature, errMsg)
							queueUpdateDraw(app, func() {
								fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler beim Ingest für %s: %s[white]\n\n", currentJob.Signature, errMsg)
								modalText.ScrollToEnd()
							})
						} else {
							logger.Error().Msgf("Failed to ingest OCFL container for %s (exit code: %d)", currentJob.Signature, exitCode)
							queueUpdateDraw(app, func() {
								fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler beim Ingest für %s (Exit-Code: %d)[white]\n\n", currentJob.Signature, exitCode)
								modalText.ScrollToEnd()
							})
						}
						if err := job.HandleIngestError(conf, currentJob, cmd.String(), output, resultJsonPath); err != nil {
							logger.Error().Err(err).Msgf("Failed to handle ingest error for %s", currentJob.Signature)
						}
						errorCount++
						continue
					}

					logger.Info().Msgf("OCFL ingest for %s completed successfully", currentJob.Signature)
					if err := job.ArchiveJob(conf, currentJob, resultJsonPath); err != nil {
						logger.Error().Err(err).Msgf("Failed to archive job for %s", currentJob.Signature)
						queueUpdateDraw(app, func() {
							fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler beim Archivieren für %s: %v[white]\n\n", currentJob.Signature, err)
							modalText.ScrollToEnd()
						})
						errorCount++
					} else {
						logger.Info().Msgf("Successfully archived job artifacts for %s", currentJob.Signature)
						queueUpdateDraw(app, func() {
							fmt.Fprintf(tview.ANSIWriter(modalText), "[green]Job %s erfolgreich archiviert.[white]\n\n", currentJob.Signature)
							modalText.ScrollToEnd()
						})
						successCount++
					}
				}

				queueUpdateDraw(app, func() {
					fmt.Fprintf(tview.ANSIWriter(modalText), "\n[green]Batch-Verarbeitung abgeschlossen.[white] Verarbeitet: %d (Erfolgreich: [green]%d[white], Fehlerhaft: [red]%d[white])\n", len(eligibleJobs), successCount, errorCount)
					modalText.ScrollToEnd()
				})
			}
		}

		queueUpdateDraw(app, func() {
			fmt.Fprintf(tview.ANSIWriter(modalText), "\n[green]Process finished. Press [yellow]ESC[green] or [yellow]ENTER[green] to close...[white]\n")
			modalText.ScrollToEnd()

			modalText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					pages.RemovePage("execModal")
					if onFinish != nil {
						onFinish()
					}
					return nil
				}
				return event
			})
		})
	}()
}
