package main

import (
	"bufio"
	"fmt"
	"io"
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

	var actionButtons []*tview.Button
	var currentJobs []Job

	var refreshCurrentBatch func()

	pages := tview.NewPages()

	// updateDetailAndActions refreshes the detail pane and dynamically adds the appropriate
	// action buttons ("OCFL erstellen", "Report erstellen", "Validate") based on the current job's state.
	updateDetailAndActions := func(job Job) {
		showDetail(job, detailView)
		buttonFlex.Clear()
		actionButtons = nil

		if job.signature == "" {
			return
		}

		if job.ocflFile == "" {
			// OCFL archive is missing: provide button to create OCFL container
			btn := tview.NewButton("OCFL erstellen").SetSelectedFunc(func() {
				createOCFL(app, pages, conf, logger, job, func() {
					if refreshCurrentBatch != nil {
						refreshCurrentBatch()
					}
					app.SetFocus(jobList)
				})
			})
			actionButtons = append(actionButtons, btn)
		} else {
			// OCFL archive exists: provide button to create report if missing, and always provide Validate button
			if job.reportFile == "" {
				btnReport := tview.NewButton("Report erstellen").SetSelectedFunc(func() {
					createReport(app, pages, conf, logger, job, func() {
						if refreshCurrentBatch != nil {
							refreshCurrentBatch()
						}
						app.SetFocus(jobList)
					})
				})
				actionButtons = append(actionButtons, btnReport)
			}

			btnValidate := tview.NewButton("Validate").SetSelectedFunc(func() {
				validateOCFL(app, pages, conf, logger, job, func() {
					if refreshCurrentBatch != nil {
						refreshCurrentBatch()
					}
					app.SetFocus(jobList)
				})
			})
			actionButtons = append(actionButtons, btnValidate)
		}

		if len(actionButtons) > 0 {
			buttonFlex.AddItem(nil, 0, 1, false)
			for i, btn := range actionButtons {
				if i > 0 {
					buttonFlex.AddItem(nil, 2, 0, false)
				}
				buttonFlex.AddItem(btn, 20, 0, false)
			}
			buttonFlex.AddItem(nil, 0, 1, false)
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
		actionButtons = nil
		currentJobs = nil
		if batchName == "" {
			return
		}
		inputFolder := conf.GetInputFolder(batchName)
		ocflFolder := conf.GetOcflFolder(batchName)
		reportFolder := conf.GetReportFolder(batchName)
		jobs, err := getSignatures(batchName, inputFolder, ocflFolder, reportFolder, logger)
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
	if conf.Batches != "" {
		batches, err := getBatches(conf.Batches)
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to read batches from %s", conf.Batches)
		} else {
			for _, b := range batches {
				batchList.AddItem(b, "", 0, nil)
			}
			if len(batches) > 0 {
				loadJobsForBatch(batches[0], 0)
			}
		}
	} else {
		logger.Warn().Msg("No batches folder specified in config")
	}

	// Upper layout section: Batches (left) | Jobs (middle) | Details + Actions (right)
	upperFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(batchList, 0, 2, true).
		AddItem(jobList, 0, 3, false).
		AddItem(detailFlex, 0, 5, false)

	// Helper to shift focus to the first action button (if present) or the detail text view
	focusDetail := func() {
		if len(actionButtons) > 0 {
			app.SetFocus(actionButtons[0])
		} else {
			app.SetFocus(detailView)
		}
	}

	// Helper to determine if focus is currently within the right detail pane
	isDetailFocused := func() bool {
		f := app.GetFocus()
		if f == detailView {
			return true
		}
		for _, btn := range actionButtons {
			if f == btn {
				return true
			}
		}
		return false
	}

	// Global key interceptor for seamless Tab, Shift-Tab, and Arrow key navigation across panels, and Ctrl+Q/Ctrl+C to quit
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if pages.HasPage("execModal") {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlQ, tcell.KeyCtrlC:
			app.Stop()
			return nil

		case tcell.KeyTab:
			if app.GetFocus() == batchList {
				app.SetFocus(jobList)
			} else if app.GetFocus() == jobList {
				focusDetail()
			} else {
				// If focused on an action button and there is a next action button, advance focus
				advanced := false
				for i, btn := range actionButtons {
					if app.GetFocus() == btn && i < len(actionButtons)-1 {
						app.SetFocus(actionButtons[i+1])
						advanced = true
						break
					}
				}
				if !advanced {
					app.SetFocus(batchList)
				}
			}
			return nil

		case tcell.KeyBacktab:
			if isDetailFocused() {
				// If focused on an action button (not the first one), move to the previous button
				movedPrev := false
				for i, btn := range actionButtons {
					if app.GetFocus() == btn && i > 0 {
						app.SetFocus(actionButtons[i-1])
						movedPrev = true
						break
					}
				}
				if !movedPrev {
					app.SetFocus(jobList)
				}
			} else if app.GetFocus() == jobList {
				app.SetFocus(batchList)
			} else {
				if len(actionButtons) > 0 {
					app.SetFocus(actionButtons[len(actionButtons)-1])
				} else {
					app.SetFocus(detailView)
				}
			}
			return nil

		case tcell.KeyRight:
			if app.GetFocus() == batchList {
				app.SetFocus(jobList)
				return nil
			} else if app.GetFocus() == jobList {
				focusDetail()
				return nil
			} else {
				for i, btn := range actionButtons {
					if app.GetFocus() == btn && i < len(actionButtons)-1 {
						app.SetFocus(actionButtons[i+1])
						return nil
					}
				}
			}

		case tcell.KeyLeft:
			for i, btn := range actionButtons {
				if app.GetFocus() == btn && i > 0 {
					app.SetFocus(actionButtons[i-1])
					return nil
				}
			}
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

	// Footer bar displaying shortcut hints
	footerView := tview.NewTextView().
		SetDynamicColors(true).
		SetText(" [yellow]Tab/Pfeiltasten[white]: Navigieren  |  [yellow]Ctrl+Q[white]: Beenden")

	// Log viewer pane at the bottom
	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	logView.SetBorder(true).SetTitle("Logs")

	// Main split layout: Upper panels on top, Logs panel in the middle, Footer bar at the bottom
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(upperFlex, 0, 2, true).
		AddItem(logView, 0, 1, false).
		AddItem(footerView, 1, 0, false)

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
