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
)

// createOCFL displays a modal execution dialog and runs the background process
// to build the OCFL container for the selected job, streaming process output in real-time.
func createOCFL(app *tview.Application, pages *tview.Pages, conf *WBConfig, job Job, onFinish func()) {
	// Create the scrollable text view to display process stdout/stderr
	modalText := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	modalText.SetBorder(true).SetTitle(fmt.Sprintf(" OCFL Erstellung: %s ", job.signature))

	// Construct a centered layout for the modal dialog box
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

	// Run external command in a background goroutine to avoid blocking the TUI event loop
	go func() {
		fmt.Fprintf(tview.ANSIWriter(modalText), "[yellow]Starte OCFL-Erstellung für %s...[white]\n\n", job.signature)

		cmd := exec.Command("gocfl",
			"create",
			filepath.Join(conf.Ocfl, fmt.Sprintf("%s.zip", strings.TrimSuffix(filepath.Base(job.infoFile), ".json"))),
			job.dataFolder,
			fmt.Sprintf("metadata:%s", job.metadataFolder),
			"-i", job.signature,
			"--ext-NNNN-metafile-source", job.infoFile)

		fmt.Fprintln(tview.ANSIWriter(modalText), strings.Join(cmd.Args, " "))

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(tview.ANSIWriter(modalText), "[red]Fehler beim Starten des Prozesses: %v[white]\n", err)
			})
		} else {
			// Stream combined stdout and stderr lines into the modal view
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

		// Prompt user upon process completion and configure key capture to dismiss the modal
		app.QueueUpdateDraw(func() {
			fmt.Fprintf(tview.ANSIWriter(modalText), "\n[green]Prozess beendet. Drücken Sie [yellow]ESC[green], [yellow]ENTER[green] oder [yellow]q[green] zum Schließen...[white]\n")
			modalText.ScrollToEnd()

			modalText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					pages.RemovePage("execModal")
					if onFinish != nil {
						onFinish()
					}
					return nil
				case tcell.KeyRune:
					if event.Rune() == 'q' || event.Rune() == 'Q' {
						pages.RemovePage("execModal")
						if onFinish != nil {
							onFinish()
						}
						return nil
					}
				}
				return event
			})
		})
	}()
}

// createReport is a placeholder for generating the PDF report for the selected job.
func createReport(app *tview.Application, pages *tview.Pages, conf *WBConfig, job Job, onFinish func()) {

}
