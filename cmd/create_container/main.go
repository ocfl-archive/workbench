package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"slices"
	"time"

	"emperror.dev/errors"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

var banner = `
██████╗  ██████╗███████╗██╗     
██╔═══██╗██╔════╝██╔════╝██║     
██║   ██║██║     █████╗  ██║     
██║   ██║██║     ██╔══╝  ██║     
╚██████╔╝╚██████╗██║     ███████╗
 ╚═════╝  ╚═════╝╚═╝     ╚══════╝
███╗   ██╗ █████╗ ████████╗██╗██╗   ██╗███████╗    
████╗  ██║██╔══██╗╚══██╔══╝██║██║   ██║██╔════╝    
██╔██╗ ██║███████║   ██║   ██║██║   ██║█████╗      
██║╚██╗██║██╔══██║   ██║   ██║╚██╗ ██╔╝██╔══╝      
██║ ╚████║██║  ██║   ██║   ██║ ╚████╔╝ ███████╗    
╚═╝  ╚═══╝╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═══╝  ╚══════╝    
 █████╗ ██████╗  ██████╗██╗  ██╗██╗██╗   ██╗███████╗
██╔══██╗██╔══██╗██╔════╝██║  ██║██║██║   ██║██╔════╝
███████║██████╔╝██║     ███████║██║██║   ██║█████╗  
██╔══██║██╔══██╗██║     ██╔══██║██║╚██╗ ██╔╝██╔══╝  
██║  ██║██║  ██║╚██████╗██║  ██║██║ ╚████╔╝ ███████╗
╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝  ╚══════╝
`

var inputFolder = flag.String("input-folder", "", "The folder to create the OCFL container in")
var ocflFolder = flag.String("ocfl-folder", "", "The folder to create the OCFL container in")
var reportFolder = flag.String("report-folder", "", "The folder to create the Report in")

func getBatches(folder string) ([]string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", folder)
	}
	var result = make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		result = append(result, entry.Name())
	}
	return result, nil
}

type Job struct {
	infoFile       string
	dataFolder     string
	metadataFolder string
	ocflFile       string
	reportFile     string
	signature      string
	title          string
}

type info struct {
	Signature string `json:"signature"`
	Title     string `json:"title"`
}

func createOCFL(app *tview.Application, pages *tview.Pages, job Job, onFinish func()) {
	modalText := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	modalText.SetBorder(true).SetTitle(fmt.Sprintf(" OCFL Erstellung: %s ", job.signature))

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
		fmt.Fprintf(tview.ANSIWriter(modalText), "[yellow]Starte OCFL-Erstellung für %s...[white]\n\n", job.signature)

		cmd := exec.Command("cmd.exe", "/c", fmt.Sprintf("echo Erstelle OCFL Container für %s... & echo InfoFile: %s", job.signature, job.infoFile))

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
			fmt.Fprintf(tview.ANSIWriter(modalText), "\n[green]Prozess beendet. Drücken Sie eine beliebige Taste zum Schließen...[white]\n")
			modalText.ScrollToEnd()

			modalText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				pages.RemovePage("execModal")
				if onFinish != nil {
					onFinish()
				}
				return nil
			})
		})
	}()
}

func createReport(app *tview.Application, pages *tview.Pages, job Job, onFinish func()) {

}

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

func getSignatures(inputF, ocflF, reportF string, logger zerolog.Logger) ([]Job, error) {
	entries, err := os.ReadDir(inputF)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", inputF)
	}
	var result = make([]Job, 0)
	var folders = make([]string, 0)
	var jsons = make([]string, 0)
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
			//			dataFolder:     path.Join(inputF, base, "data"),
			//			metadataFolder: path.Join(inputF, base, "metadata"),
		}
		if _, err := os.Stat(path.Join(inputF, base, "data")); err == nil {
			job.dataFolder = path.Join(inputF, base, "data")
		}
		if _, err := os.Stat(path.Join(inputF, base, "metadata")); err == nil {
			job.metadataFolder = path.Join(inputF, base, "metadata")
		}
		if _, err := os.Stat(path.Join(reportF, base+".pdf")); err == nil {
			job.reportFile = path.Join(reportF, base+".pdf")
		}
		if _, err := os.Stat(path.Join(ocflF, base+".zip")); err == nil {
			job.ocflFile = path.Join(ocflF, base+".zip")
		}
		result = append(result, job)
	}
	return result, nil
}

func main() {
	flag.Parse()

	errRead, errWrite, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: errWrite, TimeFormat: time.DateTime}).With().Timestamp().Logger()

	app := tview.NewApplication()

	bannerView := tview.NewTextView().
		SetText(banner).
		SetTextAlign(tview.AlignCenter)

	batchList := tview.NewList().ShowSecondaryText(false)
	batchList.SetBorder(true).SetTitle("Batches")

	jobList := tview.NewList().ShowSecondaryText(false)
	jobList.SetBorder(true).SetTitle("Jobs")

	detailView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	detailView.SetBorder(true).SetTitle("Detail")

	buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	detailFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(detailView, 0, 1, false).
		AddItem(buttonFlex, 3, 0, false)

	var actionButton *tview.Button
	var currentJobs []Job

	var refreshCurrentBatch func()

	pages := tview.NewPages()

	updateDetailAndActions := func(job Job) {
		showDetail(job, detailView)
		buttonFlex.Clear()
		actionButton = nil

		if job.signature == "" {
			return
		}

		if job.ocflFile == "" {
			btn := tview.NewButton("OCFL erstellen").SetSelectedFunc(func() {
				createOCFL(app, pages, job, func() {
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
			btn := tview.NewButton("Report erstellen").SetSelectedFunc(func() {
				createReport(app, pages, job, func() {
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
		if batchName == "" || *inputFolder == "" {
			return
		}
		batchPath := path.Join(*inputFolder, batchName)
		jobs, err := getSignatures(batchPath, *ocflFolder, *reportFolder, logger)
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

	refreshCurrentBatch = func() {
		if index := batchList.GetCurrentItem(); index >= 0 {
			batchName, _ := batchList.GetItemText(index)
			jobIndex := jobList.GetCurrentItem()
			loadJobsForBatch(batchName, jobIndex)
		}
	}

	jobList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(currentJobs) {
			updateDetailAndActions(currentJobs[index])
		}
	})

	batchList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		loadJobsForBatch(mainText, 0)
	})

	if *inputFolder != "" {
		batches, err := getBatches(*inputFolder)
		if err != nil {
			logger.Error().Err(err).Msgf("Failed to read batches from %s", *inputFolder)
		} else {
			for _, b := range batches {
				batchList.AddItem(b, "", 0, nil)
			}
			if len(batches) > 0 {
				loadJobsForBatch(batches[0], 0)
			}
		}
	} else {
		logger.Warn().Msg("No -input-folder specified")
	}

	upperFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(batchList, 0, 2, true).
		AddItem(jobList, 0, 3, false).
		AddItem(detailFlex, 0, 5, false)

	focusDetail := func() {
		if actionButton != nil {
			app.SetFocus(actionButton)
		} else {
			app.SetFocus(detailView)
		}
	}

	isDetailFocused := func() bool {
		f := app.GetFocus()
		return f == detailView || (actionButton != nil && f == actionButton)
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if pages.HasPage("execModal") {
			return event
		}

		if event.Key() == tcell.KeyTab {
			if app.GetFocus() == batchList {
				app.SetFocus(jobList)
			} else if app.GetFocus() == jobList {
				focusDetail()
			} else {
				app.SetFocus(batchList)
			}
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			if isDetailFocused() {
				app.SetFocus(jobList)
			} else if app.GetFocus() == jobList {
				app.SetFocus(batchList)
			} else {
				focusDetail()
			}
			return nil
		}
		if event.Key() == tcell.KeyRight {
			if app.GetFocus() == batchList {
				app.SetFocus(jobList)
				return nil
			} else if app.GetFocus() == jobList {
				focusDetail()
				return nil
			}
		}
		if event.Key() == tcell.KeyLeft {
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

	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	logView.SetBorder(true).SetTitle("Logs")

	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(upperFlex, 0, 2, true).
		AddItem(logView, 0, 1, false)

	pages.AddPage("main", mainFlex, true, true)

	go func() {
		scanner := bufio.NewScanner(errRead)
		for scanner.Scan() {
			line := scanner.Text()
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(tview.ANSIWriter(logView), "%s\n", line)
				logView.ScrollToEnd()
			})
		}
	}()

	time.AfterFunc(2*time.Second, func() {
		app.QueueUpdateDraw(func() {
			app.SetRoot(pages, true)
		})
	})

	if err := app.SetRoot(bannerView, true).Run(); err != nil {
		panic(err)
	}
}
