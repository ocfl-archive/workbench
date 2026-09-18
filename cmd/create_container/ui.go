package main

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

// showDetail renders formatted metadata and file path status for the given job into the target TextView.
func showDetail(job Job, target *tview.TextView) {
	target.Clear()
	fmt.Fprintf(target, "[yellow]Signature:[white] %s\n", job.signature)
	fmt.Fprintf(target, "[yellow]Title:[white]     %s\n", job.title)
	if job.infoFile != "" {
		fmt.Fprintf(target, "\n[green]Info File:[white]       %s\n", job.infoFile)
	}
	if job.dataFolder != "" {
		fmt.Fprintf(target, "[green]Data Folder:[white]     %s\n", job.dataFolder)
	}
	if job.metadataFolder != "" {
		fmt.Fprintf(target, "[green]Metadata Folder:[white] %s\n", job.metadataFolder)
	}
	if job.reportFile != "" {
		fmt.Fprintf(target, "[green]Report File:[white]     %s\n", job.reportFile)
	}
	if job.ocflFile != "" {
		fmt.Fprintf(target, "[green]OCFL File:[white]       %s\n", job.ocflFile)
	}
}

// setupUI constructs and wires all TUI components, pages, keyboard navigation handlers,
// and background log streaming routines, returning the initial banner view and pages manager.
func setupUI(app *tview.Application, conf *WBConfig, logger zerolog.Logger, logReader io.Reader) (*tview.TextView, *tview.Pages) {
	// Initialize startup banner view
	bannerView := tview.NewTextView().
		SetText(banner).
		SetTextAlign(tview.AlignCenter)

	// Batches list on the left panel
	batchList := tview.NewList().ShowSecondaryText(false)
	batchList.SetBorder(true).SetTitle("Batches")

	// Jobs list in the middle panel
	jobList := tview.NewList().ShowSecondaryText(false)
	jobList.SetBorder(true).SetTitle("Jobs")

	// Detail view displaying selected job metadata on the right panel
	detailView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	detailView.SetBorder(true).SetTitle("Detail")

	// Container for dynamic action buttons below the detail view
	buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	detailFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(detailView, 0, 1, false).
		AddItem(buttonFlex, 3, 0, false)

	var actionButton *tview.Button
	var currentJobs []Job

	var refreshCurrentBatch func()

	pages := tview.NewPages()

	// updateDetailAndActions refreshes the detail pane and dynamically adds the appropriate
	// action button ("OCFL erstellen" or "Report erstellen") based on the current job's state.
	updateDetailAndActions := func(job Job) {
		showDetail(job, detailView)
		buttonFlex.Clear()
		actionButton = nil

		if job.signature == "" {
			return
		}

		if job.ocflFile == "" {
			// OCFL archive is missing: provide button to create OCFL container
			btn := tview.NewButton("OCFL erstellen").SetSelectedFunc(func() {
				createOCFL(app, pages, conf, job, func() {
					if refreshCurrentBatch != nil {
						refreshCurrentBatch()
					}
					app.SetFocus(jobList)
				})
			})
			actionButton = btn
			buttonFlex.AddItem(nil, 0, 1, false).
				AddItem(btn, 20, 0, false).
				AddItem(nil, 0, 1, false)
		} else if job.reportFile == "" {
			// OCFL archive exists but PDF report is missing: provide button to create report
			btn := tview.NewButton("Report erstellen").SetSelectedFunc(func() {
				createReport(app, pages, conf, job, func() {
					if refreshCurrentBatch != nil {
						refreshCurrentBatch()
					}
					app.SetFocus(jobList)
				})
			})
			actionButton = btn
			buttonFlex.AddItem(nil, 0, 1, false).
				AddItem(btn, 20, 0, false).
				AddItem(nil, 0, 1, false)
		}
	}

	// loadJobsForBatch scans the selected batch directory and populates the jobs list,
	// optionally preserving the selected job index across refreshes.
	loadJobsForBatch := func(batchName string, preserveIndex ...int) {
		targetIndex := 0
		if len(preserveIndex) > 0 && preserveIndex[0] >= 0 {
			targetIndex = preserveIndex[0]
		} else if currentIdx := jobList.GetCurrentItem(); currentIdx >= 0 {
			targetIndex = currentIdx
		}

		jobList.Clear()
		detailView.Clear()
		buttonFlex.Clear()
		actionButton = nil
		currentJobs = nil
		if batchName == "" || conf.Input == "" {
			return
		}
		batchPath := path.Join(conf.Input, batchName)
		jobs, err := getSignatures(batchPath, conf.Ocfl, conf.Report, logger)
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to get jobs for batch %s", batchName)
			return
		}
		currentJobs = jobs
		for _, job := range jobs {
			jobList.AddItem(job.signature, "", 0, nil)
		}
		if len(currentJobs) > 0 {
			if targetIndex >= len(currentJobs) {
				targetIndex = len(currentJobs) - 1
			}
			jobList.SetCurrentItem(targetIndex)
			updateDetailAndActions(currentJobs[targetIndex])
		}
	}

	// refreshCurrentBatch re-evaluates the jobs of the currently selected batch while maintaining the selection index
	refreshCurrentBatch = func() {
		if index := batchList.GetCurrentItem(); index >= 0 {
			batchName, _ := batchList.GetItemText(index)
			jobIndex := jobList.GetCurrentItem()
			loadJobsForBatch(batchName, jobIndex)
		}
	}

	// Update detail view and action buttons when a different job is highlighted
	jobList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(currentJobs) {
			updateDetailAndActions(currentJobs[index])
		}
	})

	// Reload jobs when a different batch is selected
	batchList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		loadJobsForBatch(mainText, 0)
	})

	// Initial population of the batch list
	if conf.Input != "" {
		batches, err := getBatches(conf.Input)
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to read batches from %s", conf.Input)
		} else {
			for _, b := range batches {
				batchList.AddItem(b, "", 0, nil)
			}
			if len(batches) > 0 {
				loadJobsForBatch(batches[0], 0)
			}
		}
	} else {
		logger.Warn().Msg("No input specified in config")
	}

	// Upper layout section: Batches (left) | Jobs (middle) | Details + Actions (right)
	upperFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(batchList, 0, 2, true).
		AddItem(jobList, 0, 3, false).
		AddItem(detailFlex, 0, 5, false)

	// Helper to shift focus to the action button (if present) or the detail text view
	focusDetail := func() {
		if actionButton != nil {
			app.SetFocus(actionButton)
		} else {
			app.SetFocus(detailView)
		}
	}

	// Helper to determine if focus is currently within the right detail pane
	isDetailFocused := func() bool {
		f := app.GetFocus()
		return f == detailView || (actionButton != nil && f == actionButton)
	}

	// Global key interceptor for seamless Tab, Shift-Tab, and Arrow key navigation across panels
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if pages.HasPage("execModal") {
			return event
		}

		if event.Rune() == 'q' || event.Rune() == 'Q' {
			app.Stop()
			return nil
		}

		switch event.Key() {
		case tcell.KeyTab:
			if app.GetFocus() == batchList {
				app.SetFocus(jobList)
			} else if app.GetFocus() == jobList {
				focusDetail()
			} else {
				app.SetFocus(batchList)
			}
			return nil

		case tcell.KeyBacktab:
			if isDetailFocused() {
				app.SetFocus(jobList)
			} else if app.GetFocus() == jobList {
				app.SetFocus(batchList)
			} else {
				focusDetail()
			}
			return nil

		case tcell.KeyRight:
			if app.GetFocus() == batchList {
				app.SetFocus(jobList)
				return nil
			} else if app.GetFocus() == jobList {
				focusDetail()
				return nil
			}

		case tcell.KeyLeft:
			if isDetailFocused() {
				app.SetFocus(jobList)
				return nil
			} else if app.GetFocus() == jobList {
				app.SetFocus(batchList)
				return nil
			}
		}

		return event
	})

	// Log viewer pane at the bottom
	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	logView.SetBorder(true).SetTitle("Logs")

	// Main split layout: Upper panels on top, Logs panel at the bottom
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(upperFlex, 0, 2, true).
		AddItem(logView, 0, 1, false)

	pages.AddPage("main", mainFlex, true, true)

	// Asynchronous reader streaming logs from the pipe into the log view with ANSI color support
	go func() {
		scanner := bufio.NewScanner(logReader)
		for scanner.Scan() {
			line := scanner.Text()
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(tview.ANSIWriter(logView), "%s\n", line)
				logView.ScrollToEnd()
			})
		}
	}()

	// Transition from the startup banner to the main application interface after 2 seconds
	time.AfterFunc(2*time.Second, func() {
		app.QueueUpdateDraw(func() {
			app.SetRoot(pages, true)
		})
	})

	return bannerView, pages
}
