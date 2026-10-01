package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureTargetDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Case 1: Directory already exists
	existingDir := filepath.Join(tempDir, "existing")
	if err := os.Mkdir(existingDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := EnsureTargetDirectory(existingDir); err != nil {
		t.Errorf("EnsureTargetDirectory(existingDir) returned unexpected error: %v", err)
	}

	// Case 2: Leaf directory missing, parent exists -> should succeed and create
	leafDir := filepath.Join(tempDir, "new_leaf")
	if err := EnsureTargetDirectory(leafDir); err != nil {
		t.Errorf("EnsureTargetDirectory(leafDir) returned unexpected error: %v", err)
	}
	if fi, err := os.Stat(leafDir); err != nil || !fi.IsDir() {
		t.Errorf("EnsureTargetDirectory(leafDir) did not create directory")
	}

	// Case 3: Parent and leaf missing -> should fail
	deepDir := filepath.Join(tempDir, "non_existent_parent", "target")
	if err := EnsureTargetDirectory(deepDir); err == nil {
		t.Errorf("EnsureTargetDirectory(deepDir) expected error for missing parent, got nil")
	}

	// Case 4: Target exists but is a file
	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := EnsureTargetDirectory(filePath); err == nil {
		t.Errorf("EnsureTargetDirectory(filePath) expected error for regular file, got nil")
	}
}

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("GetFreePort() returned error: %v", err)
	}
	if port <= 0 || port > 65535 {
		t.Errorf("GetFreePort() returned invalid port number: %d", port)
	}
}
