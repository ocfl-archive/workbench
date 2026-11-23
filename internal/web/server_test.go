package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ocfl-archive/workbench/internal/config"
	"github.com/rs/zerolog"
)

func setupTestServer(t *testing.T) (*Server, string) {
	tempDir := t.TempDir()
	batchName := "batch_001"

	conf := &config.WBConfig{
		Batches:    tempDir,
		Input:      filepath.Join(tempDir, "{batch}", "incoming"),
		Ocfl:       filepath.Join(tempDir, "{batch}", "ocfl"),
		Report:     filepath.Join(tempDir, "{batch}", "report"),
		Error:      filepath.Join(tempDir, "{batch}", "error"),
		Archived:   filepath.Join(tempDir, "{batch}", "archived"),
		ConfigFile: filepath.Join(tempDir, "workbench.toml"),
	}

	incomingJobDir := filepath.Join(conf.GetInputFolder(batchName), "job_test")
	if err := os.MkdirAll(incomingJobDir, 0755); err != nil {
		t.Fatal(err)
	}
	infoJSON := `{"signature":"SIG-WEB-01","title":"Web Test Archival Unit"}`
	if err := os.WriteFile(filepath.Join(incomingJobDir, "info.json"), []byte(infoJSON), 0644); err != nil {
		t.Fatal(err)
	}

	logger := zerolog.Nop()
	server, err := NewServer(ServerConfig{
		Addr:       ":0",
		Conf:       conf,
		Logger:     logger,
		ConfigFile: conf.ConfigFile,
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	return server, batchName
}

func TestServer_IndexRoute(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Workbench") {
		t.Errorf("expected body to contain 'Workbench', got %s", body)
	}
	if !strings.Contains(body, "batch_001") {
		t.Errorf("expected body to contain 'batch_001', got %s", body)
	}
	if !strings.Contains(body, "SIG-WEB-01") {
		t.Errorf("expected body to contain 'SIG-WEB-01', got %s", body)
	}
}

func TestServer_BatchesFragment(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/batches", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "batch_001") {
		t.Errorf("expected body to contain 'batch_001', got %s", body)
	}
}

func TestServer_JobsFragment(t *testing.T) {
	server, batchName := setupTestServer(t)

	req := httptest.NewRequest("GET", "/batch/"+batchName, nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "SIG-WEB-01") {
		t.Errorf("expected body to contain 'SIG-WEB-01', got %s", body)
	}
	if !strings.Contains(body, "Batch durchführen") {
		t.Errorf("expected body to contain 'Batch durchführen', got %s", body)
	}
}

func TestServer_JobDetailFragment(t *testing.T) {
	server, batchName := setupTestServer(t)

	req := httptest.NewRequest("GET", "/batch/"+batchName+"/job/job_test", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "SIG-WEB-01") {
		t.Errorf("expected body to contain 'SIG-WEB-01', got %s", body)
	}
	if !strings.Contains(body, "OCFL erstellen") {
		t.Errorf("expected body to contain action 'OCFL erstellen', got %s", body)
	}
}

func TestServer_StaticAssetDelivery(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/static/css/tailwind.min.css", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for static asset, got %d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Errorf("expected non-empty static asset content")
	}
}

func TestServer_APIStatus(t *testing.T) {
	server, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "running") {
		t.Errorf("expected json to contain 'running', got %s", body)
	}
}
