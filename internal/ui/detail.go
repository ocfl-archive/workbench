package ui

import (
	"fmt"

	"github.com/ocfl-archive/workbench/internal/job"
	"github.com/rivo/tview"
)

// ShowDetail renders formatted metadata and file path status for the given job into the target TextView.
func ShowDetail(j job.Job, target *tview.TextView) {
	target.Clear()
	fmt.Fprintf(target, "[yellow]Signature:[white] %s\n", j.Signature)
	fmt.Fprintf(target, "[yellow]Title:[white]     %s\n", j.Title)
	if j.InfoFile != "" {
		fmt.Fprintf(target, "\n[green]Info File:[white]       %s\n", j.InfoFile)
	}
	if j.DataFolder != "" {
		fmt.Fprintf(target, "[green]Data Folder:[white]     %s\n", j.DataFolder)
	}
	if j.MetadataFolder != "" {
		fmt.Fprintf(target, "[green]Metadata Folder:[white] %s\n", j.MetadataFolder)
	}
	if j.ReportFile != "" {
		fmt.Fprintf(target, "[green]Report File:[white]     %s\n", j.ReportFile)
	}
	if j.OcflFile != "" {
		fmt.Fprintf(target, "[green]OCFL File:[white]       %s\n", j.OcflFile)
	}
}
