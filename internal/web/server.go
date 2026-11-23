package web

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/ocfl-archive/workbench/internal/job"
	"github.com/ocfl-archive/workbench/internal/ui"
	"github.com/rs/zerolog"
)

// ServerConfig holds the initialization settings for the Workbench Web server.
type ServerConfig struct {
	Addr       string
	Conf       *config.WBConfig
	Logger     zerolog.Logger
	ConfigFile string
}

// Server coordinates HTTP routing, template rendering, and background execution.
type Server struct {
	addr        string
	conf        *config.WBConfig
	logger      zerolog.Logger
	broadcaster *Broadcaster
	execManager *ExecutionManager
	templates   *template.Template
	httpServer  *http.Server
	mux         *http.ServeMux
}

// PageData represents the view-model supplied to the web templates.
type PageData struct {
	Title         string
	ConfigFile    string
	BatchesFolder string
	Batches       []string
	SelectedBatch string
	Jobs          []job.Job
	SelectedJob   *job.Job
	TotalBatches  int
	IncomingCount int
	ErrorCount    int
	ArchivedCount int
	ServerAddr    string
	Banner        string
	IsRunning     bool
	ActiveTask    string
	History       []LogEntry
}

// NewServer initializes the Server instance and compiles embedded templates.
func NewServer(cfg ServerConfig) (*Server, error) {
	broadcaster := NewBroadcaster(300)
	execManager := NewExecutionManager(cfg.Conf, broadcaster, cfg.Logger)

	// Custom template helper functions
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires even number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"hasAction": func(actions []job.ActionType, action string) bool {
			for _, a := range actions {
				if string(a) == action {
					return true
				}
			}
			return false
		},
		"actionLabel": func(action string) string {
			return job.ActionType(action).Label()
		},
		"formatTime": func(t time.Time) string {
			return t.Format("15:04:05")
		},
		"fileName": func(path string) string {
			return filepath.Base(path)
		},
		"fileExists": func(path string) bool {
			if path == "" {
				return false
			}
			_, err := os.Stat(path)
			return err == nil
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(ContentFS, "templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse web templates: %w", err)
	}

	s := &Server{
		addr:        cfg.Addr,
		conf:        cfg.Conf,
		logger:      cfg.Logger,
		broadcaster: broadcaster,
		execManager: execManager,
		templates:   tmpl,
		mux:         http.NewServeMux(),
	}

	s.setupRoutes()
	return s, nil
}

// Broadcaster returns the server's SSE broadcaster.
func (s *Server) Broadcaster() *Broadcaster {
	return s.broadcaster
}

func (s *Server) setupRoutes() {
	// 1. Static asset delivery from embedded filesystem
	subStatic, err := fs.Sub(ContentFS, "static")
	if err == nil {
		s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(subStatic))))
	}

	// 2. Main Dashboard & Fragment routes
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /batches", s.handleBatchesFragment)
	s.mux.HandleFunc("GET /batch/{batch}", s.handleJobsFragment)
	s.mux.HandleFunc("GET /batch/{batch}/job/{basename}", s.handleJobDetailFragment)

	// 3. Action & Execution routes
	s.mux.HandleFunc("POST /batch/{batch}/job/{basename}/action/{action}", s.handleJobAction)
	s.mux.HandleFunc("POST /batch/{batch}/run", s.handleBatchRun)

	// 4. Real-time Server-Sent Events stream
	s.mux.HandleFunc("GET /events", s.handleSSE)

	// 5. REST Status endpoint
	s.mux.HandleFunc("GET /api/status", s.handleAPIStatus)
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start() error {
	if s.addr == "" {
		s.addr = ":8765"
	}

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // Disabled for SSE streaming
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info().Msgf("Workbench Web GUI running on http://%s", s.addr)
	s.broadcaster.AppendLog("info", fmt.Sprintf("Workbench Web GUI started on http://%s", s.addr))

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully stops the web server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) getBatches() []string {
	batchesFolder := s.conf.Batches
	if batchesFolder == "" {
		batchesFolder = "."
	}
	batches, err := job.GetBatches(batchesFolder)
	if err != nil {
		s.logger.Warn().Err(err).Msgf("Failed to scan batches in %s", batchesFolder)
		return nil
	}
	return batches
}

func (s *Server) getJobsForBatch(batchName string) []job.Job {
	if batchName == "" {
		return nil
	}
	inputFolder := s.conf.GetInputFolder(batchName)
	ocflFolder := s.conf.GetOcflFolder(batchName)
	reportFolder := s.conf.GetReportFolder(batchName)
	errorFolder := s.conf.GetErrorFolder(batchName)
	archivedFolder := s.conf.GetArchivedFolder(batchName)

	jobs, err := job.GetSignatures(batchName, inputFolder, ocflFolder, reportFolder, errorFolder, archivedFolder, s.logger)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Failed to get jobs for batch %s", batchName)
		return nil
	}
	return jobs
}

func (s *Server) findJob(batchName, baseName string) *job.Job {
	jobs := s.getJobsForBatch(batchName)
	for _, j := range jobs {
		if j.BaseName == baseName {
			return &j
		}
	}
	return nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	batches := s.getBatches()
	selectedBatch := r.URL.Query().Get("batch")
	if selectedBatch == "" && len(batches) > 0 {
		selectedBatch = batches[0]
	}

	var jobs []job.Job
	var selectedJob *job.Job
	var incomingCount, errorCount, archivedCount int

	if selectedBatch != "" {
		jobs = s.getJobsForBatch(selectedBatch)
		for _, j := range jobs {
			switch j.Status {
			case job.StatusIncoming:
				incomingCount++
			case job.StatusError:
				errorCount++
			case job.StatusArchived:
				archivedCount++
			}
		}

		selectedBaseName := r.URL.Query().Get("job")
		if selectedBaseName != "" {
			selectedJob = s.findJob(selectedBatch, selectedBaseName)
		}
		if selectedJob == nil && len(jobs) > 0 {
			selectedJob = &jobs[0]
		}
	}

	isRunning, activeTask := s.execManager.IsRunning()

	data := PageData{
		Title:         "Workbench Web GUI",
		ConfigFile:    s.conf.ConfigFile,
		BatchesFolder: s.conf.Batches,
		Batches:       batches,
		SelectedBatch: selectedBatch,
		Jobs:          jobs,
		SelectedJob:   selectedJob,
		TotalBatches:  len(batches),
		IncomingCount: incomingCount,
		ErrorCount:    errorCount,
		ArchivedCount: archivedCount,
		ServerAddr:    s.addr,
		Banner:        ui.Banner,
		IsRunning:     isRunning,
		ActiveTask:    activeTask,
		History:       s.broadcaster.GetHistory(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		s.logger.Error().Err(err).Msg("Failed to execute index template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleBatchesFragment(w http.ResponseWriter, r *http.Request) {
	batches := s.getBatches()
	selectedBatch := r.URL.Query().Get("batch")
	if selectedBatch == "" && len(batches) > 0 {
		selectedBatch = batches[0]
	}

	data := map[string]interface{}{
		"Batches":       batches,
		"SelectedBatch": selectedBatch,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "batches.gohtml", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleJobsFragment(w http.ResponseWriter, r *http.Request) {
	batchName, err := url.PathUnescape(r.PathValue("batch"))
	if err != nil || batchName == "" {
		http.Error(w, "Ungültiger Batch", http.StatusBadRequest)
		return
	}

	jobs := s.getJobsForBatch(batchName)
	var incomingCount, errorCount, archivedCount int
	for _, j := range jobs {
		switch j.Status {
		case job.StatusIncoming:
			incomingCount++
		case job.StatusError:
			errorCount++
		case job.StatusArchived:
			archivedCount++
		}
	}

	var selectedJob *job.Job
	if len(jobs) > 0 {
		selectedJob = &jobs[0]
	}

	data := map[string]interface{}{
		"Batch":         batchName,
		"Jobs":          jobs,
		"SelectedJob":   selectedJob,
		"IncomingCount": incomingCount,
		"ErrorCount":    errorCount,
		"ArchivedCount": archivedCount,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "jobs.gohtml", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleJobDetailFragment(w http.ResponseWriter, r *http.Request) {
	batchName, err := url.PathUnescape(r.PathValue("batch"))
	if err != nil || batchName == "" {
		http.Error(w, "Ungültiger Batch", http.StatusBadRequest)
		return
	}
	baseName, err := url.PathUnescape(r.PathValue("basename"))
	if err != nil || baseName == "" {
		http.Error(w, "Ungültiger Job", http.StatusBadRequest)
		return
	}

	j := s.findJob(batchName, baseName)
	if j == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="p-8 text-center text-base-content/60 text-sm">Job nicht gefunden</div>`)
		return
	}

	data := map[string]interface{}{
		"Job": j,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "detail.gohtml", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleJobAction(w http.ResponseWriter, r *http.Request) {
	batchName, _ := url.PathUnescape(r.PathValue("batch"))
	baseName, _ := url.PathUnescape(r.PathValue("basename"))
	actionStr, _ := url.PathUnescape(r.PathValue("action"))

	j := s.findJob(batchName, baseName)
	if j == nil {
		http.Error(w, "Job nicht gefunden", http.StatusNotFound)
		return
	}

	action := job.ActionType(actionStr)
	if err := s.execManager.ExecuteJobAction(*j, action); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, `<div class="alert alert-error shadow-sm text-xs py-2 px-3"><span>%s</span></div>`, escapeHTML(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "terminalOpen")
	fmt.Fprintf(w, `<div class="alert alert-info shadow-sm text-xs py-2 px-3"><span>Aktion '%s' für %s gestartet...</span></div>`, escapeHTML(action.Label()), escapeHTML(j.Signature))
}

func (s *Server) handleBatchRun(w http.ResponseWriter, r *http.Request) {
	batchName, _ := url.PathUnescape(r.PathValue("batch"))
	if batchName == "" {
		http.Error(w, "Kein Batch ausgewählt", http.StatusBadRequest)
		return
	}

	if err := s.execManager.ExecuteBatch(batchName); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, `<div class="alert alert-error shadow-sm text-xs py-2 px-3"><span>%s</span></div>`, escapeHTML(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "terminalOpen")
	fmt.Fprintf(w, `<div class="alert alert-info shadow-sm text-xs py-2 px-3"><span>Batch-Verarbeitung für '%s' gestartet...</span></div>`, escapeHTML(batchName))
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	clientChan := s.broadcaster.Subscribe()
	defer s.broadcaster.Unsubscribe(clientChan)

	// Send initial connection heartbeat
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	notify := r.Context().Done()
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			if msg.Event != "" {
				fmt.Fprintf(w, "event: %s\n", msg.Event)
			}
			// Multi-line data handling for SSE standard
			lines := strings.Split(msg.Data, "\n")
			for _, line := range lines {
				fmt.Fprintf(w, "data: %s\n", line)
			}
			fmt.Fprintf(w, "\n")
			flusher.Flush()
		}
	}
}

func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	isRunning, activeTask := s.execManager.IsRunning()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"running":%t,"task":%q}`, isRunning, activeTask)
}
