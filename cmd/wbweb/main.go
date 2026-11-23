// Package main provides the entry point for the wbweb web application.
//
// wbweb is a modern, responsive Web User Interface (Web-GUI) for archival workflows,
// built analogously to the wb TUI with Tailwind CSS, DaisyUI, HTMX 4, and real-time SSE.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/ocfl-archive/workbench/internal/web"
	"github.com/rs/zerolog"
)

var (
	configPath = flag.String("config", "", "The path to the workbench TOML configuration file")
	addr       = flag.String("addr", ":8765", "The HTTP server address (e.g. :8765 or localhost:8765)")
	openBrowser = flag.Bool("open", false, "Automatically open the Web GUI in the default browser")
)

func openURL(targetURL string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	case "darwin":
		cmd = exec.Command("open", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	_ = cmd.Start()
}

func main() {
	flag.Parse()

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Channel for graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	// Temporary standard logger before server initialization
	consoleOut := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}
	tempLogger := zerolog.New(consoleOut).With().Timestamp().Logger()

	server, err := web.NewServer(web.ServerConfig{
		Addr:       *addr,
		Conf:       conf,
		Logger:     tempLogger,
		ConfigFile: conf.ConfigFile,
	})
	if err != nil {
		tempLogger.Fatal().Err(err).Msg("Failed to initialize web server")
	}

	// Pipe log events both to stdout and to the web SSE broadcaster
	broadcasterWriter := server.Broadcaster().LogWriter("info")
	multiWriter := io.MultiWriter(consoleOut, broadcasterWriter)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()

	logger.Info().Msgf("Using config file: %s", conf.ConfigFile)
	logger.Info().Msgf("Batches folder: %s", conf.Batches)

	serverURL := *addr
	if serverURL[0] == ':' {
		serverURL = fmt.Sprintf("http://localhost%s", serverURL)
	} else if serverURL[:4] != "http" {
		serverURL = fmt.Sprintf("http://%s", serverURL)
	}

	if *openBrowser {
		go func() {
			// Small delay to ensure server has started
			time.Sleep(300 * time.Millisecond)
			openURL(serverURL)
		}()
	}

	go func() {
		if err := server.Start(); err != nil {
			logger.Error().Err(err).Msg("Web server error")
		}
	}()

	fmt.Printf("\n=======================================================\n")
	fmt.Printf(" Workbench Web GUI is running at %s\n", serverURL)
	fmt.Printf(" Press CTRL+C to stop the server\n")
	fmt.Printf("=======================================================\n\n")

	<-stopChan
	logger.Info().Msg("Shutting down Workbench Web GUI gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server shutdown error")
	} else {
		logger.Info().Msg("Server stopped successfully")
	}
}

// ensure listener address is valid
func checkAddress(address string) bool {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}
