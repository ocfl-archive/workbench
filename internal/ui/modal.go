package ui

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
	app.SetFocus(modalText)

	go func() {
		var output strings.Builder
		fmt.Fprintf(tview.ANSIWriter(modalText), "%s\n\n", startMsg)
		fmt.Fprintln(tview.ANSIWriter(modalText), strings.Join(cmd.Args, " "))

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		var exitCode int
		if err := cmd.Start(); err != nil {
			exitCode = -1
			errMsg := fmt.Sprintf("Error starting process: %v\n", err)
			output.WriteString(errMsg)
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(tview.ANSIWriter(modalText), "[red]%s[white]", errMsg)
			})
		} else {
			multiReader := io.MultiReader(stdout, stderr)
			scanner := bufio.NewScanner(multiReader)
			for scanner.Scan() {
				line := scanner.Text()
				output.WriteString(line)
				output.WriteString("\n")
				app.QueueUpdateDraw(func() {
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

		app.QueueUpdateDraw(func() {
			fmt.Fprintf(tview.ANSIWriter(modalText), "\n[green]Process finished. Press [yellow]ESC[green] or [yellow]ENTER[green] to close...[white]\n")
			modalText.ScrollToEnd()

			modalText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					pages.RemovePage("execModal")
					if onFinish != nil {
						onFinish(exitCode, output.String())
					}
					return nil
				}
				return event
			})
		})
	}()
}
