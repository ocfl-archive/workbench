package main

import (
	"flag"
	"os"
	"time"

	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

// Command-line configuration flags specifying directory paths for inputs and output artifacts.
var configPath = flag.String("config", "", "The path to the configuration file")

// main parses CLI arguments, configures log redirection to the TUI log viewer,
// and starts the application event loop.
func main() {
	flag.Parse()

	conf, err := LoadConfig(*configPath)
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

	app := tview.NewApplication().EnableMouse(true)

	// Build the TUI components, passing the configuration and the read end of the pipe for real-time log streaming
	bannerView, _ := setupUI(app, conf, logger, errRead)

	// Display the banner view on startup; setupUI will switch to the main layout after a short delay
	if err := app.SetRoot(bannerView, true).Run(); err != nil {
		panic(err)
	}
}
