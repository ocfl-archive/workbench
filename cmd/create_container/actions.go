package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

// runCommandInModal encapsulates modal dialog creation and process output streaming.
func runCommandInModal(
	app *tview.Application,
	pages *tview.Pages,
	title string,
	startMsg string,
	cmd *exec.Cmd,
	onFinish func(),
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
	app.SetFocus(modalText)

	go func() {
		fmt.Fprintf(tview.ANSIWriter(modalText), "%s\n\n", startMsg)
		fmt.Fprintln(tview.ANSIWriter(modalText), strings.Join(cmd.Args, " "))

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler beim Starten des Prozesses: %v[white]\n", err)
			})
		} else {
			multiReader := io.MultiReader(stdout, stderr)
			scanner := bufio.NewScanner(multiReader)
			for scanner.Scan() {
				line := scanner.Text()
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(tview.ANSIWriter(modalText), "%s\n", line)
					modalText.ScrollToEnd()
				})
			}
			_ = cmd.Wait()
		}

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(tview.ANSIWriter(modalText), "\n[green]Prozess beendet. Drücken Sie [yellow]ESC[green] oder [yellow]ENTER[green] zum Schließen...[white]\n")
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

// createOCFL displays a modal execution dialog and runs the background process
// to build the OCFL container for the selected job, streaming process output in real-time.
func createOCFL(app *tview.Application, pages *tview.Pages, conf *WBConfig, logger zerolog.Logger, job Job, onFinish func()) {
	zipName := fmt.Sprintf("%s.zip", job.baseName)
	cmd := exec.Command("gocfl",
		"create",
		filepath.Join(conf.Ocfl, zipName),
		job.dataFolder,
		fmt.Sprintf("metadata:%s", job.metadataFolder),
		"-i", job.signature,
		"--ext-NNNN-metafile-source", job.infoFile,
	)

	title := fmt.Sprintf(" OCFL Erstellung: %s ", job.signature)
	startMsg := fmt.Sprintf("[yellow]Starte OCFL-Erstellung für %s...[white]", job.signature)

	runCommandInModal(app, pages, title, startMsg, cmd, onFinish)
}

// createReport generates the PDF report for the selected job.
func createReport(app *tview.Application, pages *tview.Pages, conf *WBConfig, logger zerolog.Logger, job Job, onFinish func()) {
	port, err := GetFreePort()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get free network port")
		return
	}

	zipName := fmt.Sprintf("%s.zip", job.baseName)
	pdfName := fmt.Sprintf("%s.pdf", job.baseName)

	cmd := exec.Command("gocfl",
		"display",
		filepath.Join(conf.Ocfl, zipName),
		"--display-fullreport", filepath.Join(conf.Report, pdfName),
		"--display-id", job.signature,
		"-a", fmt.Sprintf("localhost:%d", port),
		"-e", fmt.Sprintf("http://localhost:%d", port),
	)

	title := fmt.Sprintf(" Report Erstellung: %s ", job.signature)
	startMsg := fmt.Sprintf("[yellow]Starte OCFL-Report-Erstellung für %s...[white]", job.signature)

	runCommandInModal(app, pages, title, startMsg, cmd, onFinish)
}

// validateOCFL displays a modal execution dialog and runs the background process
// to validate the OCFL container for the selected job, streaming process output in real-time.
func validateOCFL(app *tview.Application, pages *tview.Pages, conf *WBConfig, logger zerolog.Logger, job Job, onFinish func()) {
	zipName := fmt.Sprintf("%s.zip", job.baseName)
	cmd := exec.Command("gocfl",
		"validate",
		filepath.Join(conf.Ocfl, zipName),
	)

	title := fmt.Sprintf(" OCFL Validierung: %s ", job.signature)
	startMsg := fmt.Sprintf("[yellow]Starte OCFL-Validierung für %s...[white]", job.signature)

	runCommandInModal(app, pages, title, startMsg, cmd, onFinish)
}
