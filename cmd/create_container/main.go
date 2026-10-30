// Package main provides the entry point for the create_container CLI application.
//
// create_container is an interactive Terminal User Interface (TUI) tool designed
// for archival workflows. It scans batch directories for digitization units,
// validates metadata, and coordinates OCFL container creation, PDF report generation,
// validation, and ingest using gocfl and ona CLI tools.
package main

import (
	"flag"
	"os"
	"time"

	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/ocfl-archive/workbench/internal/ui"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

// configPath specifies the CLI flag for the path to the workbench TOML configuration file.
var configPath = flag.String("config", "", "The path to the configuration file")

// main initializes configuration, sets up real-time log streaming via an OS pipe,
// constructs the TUI application layout, and starts the event loop.
func main() {
	flag.Parse()

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		panic(err)
	}

	// Create an OS pipe to redirect zerolog output to the TUI log panel
	errRead, errWrite, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// Initialize zerolog to write formatted logs into the pipe
	logger := zerolog.New(zerolog.ConsoleWriter{Out: errWrite, TimeFormat: time.DateTime}).With().Timestamp().Logger()
	logger.Info().Msgf("Using config file: %s", conf.ConfigFile)

	app := tview.NewApplication().EnableMouse(false)

	// Build the TUI components, passing the configuration and the read end of the pipe for real-time log streaming
	bannerView, _ := ui.SetupUI(app, conf, logger, errRead)

	// Display the banner view on startup; SetupUI will switch to the main layout after a short delay
	app.SetRoot(bannerView, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
