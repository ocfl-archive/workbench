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
	if j.Status != "" {
		fmt.Fprintf(target, "[yellow]Status:[white]    %s\n", j.Status)
	}
	if j.IngestID != "" {
		fmt.Fprintf(target, "[yellow]Ingest ID:[white] %s\n", j.IngestID)
	}
	if j.IncomingFolder != "" {
		fmt.Fprintf(target, "\n[green]Incoming Folder:[white] %s\n", j.IncomingFolder)
	}
	if j.InfoFile != "" {
		if j.IncomingFolder == "" {
			fmt.Fprintf(target, "\n")
		}
		fmt.Fprintf(target, "[green]Info File:[white]       %s\n", j.InfoFile)
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
	if j.ChecksumFile != "" {
		fmt.Fprintf(target, "[green]Checksum File:[white]   %s\n", j.ChecksumFile)
	}
	if j.UploadFile != "" {
		fmt.Fprintf(target, "[green]Upload File:[white]     %s\n", j.UploadFile)
	}
	if j.ErrorFile != "" {
		fmt.Fprintf(target, "[green]Error File:[white]      %s\n", j.ErrorFile)
	}
}
