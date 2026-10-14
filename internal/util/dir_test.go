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

func TestMoveDirectory_SameVolume(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "src_job")
	dst := filepath.Join(tempDir, "dst_job")

	// Create source structure: src/info.json, src/data/file1.txt
	if err := os.MkdirAll(filepath.Join(src, "data"), 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "info.json"), []byte(`{"signature":"SIG-1"}`), 0644); err != nil {
		t.Fatalf("failed to write info.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "data", "file1.txt"), []byte("payload data"), 0644); err != nil {
		t.Fatalf("failed to write file1.txt: %v", err)
	}

	if err := MoveDirectory(src, dst); err != nil {
		t.Fatalf("MoveDirectory failed: %v", err)
	}

	// Verify source no longer exists
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("expected source '%s' to be removed, but stat returned: %v", src, err)
	}

	// Verify destination files exist
	infoBytes, err := os.ReadFile(filepath.Join(dst, "info.json"))
	if err != nil {
		t.Fatalf("failed to read destination info.json: %v", err)
	}
	if string(infoBytes) != `{"signature":"SIG-1"}` {
		t.Errorf("unexpected content in dst info.json: %s", string(infoBytes))
	}

	payloadBytes, err := os.ReadFile(filepath.Join(dst, "data", "file1.txt"))
	if err != nil {
		t.Fatalf("failed to read destination file1.txt: %v", err)
	}
	if string(payloadBytes) != "payload data" {
		t.Errorf("unexpected content in dst file1.txt: %s", string(payloadBytes))
	}
}

func TestMoveDirectory_TargetExists(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "src_job")
	dst := filepath.Join(tempDir, "dst_job")

	// Setup src with new_file.txt
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatalf("failed to create src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "new_file.txt"), []byte("new content"), 0644); err != nil {
		t.Fatalf("failed to write src file: %v", err)
	}

	// Setup dst with old_file.txt
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatalf("failed to create dst: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dst, "old_file.txt"), []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to write old dst file: %v", err)
	}

	if err := MoveDirectory(src, dst); err != nil {
		t.Fatalf("MoveDirectory failed: %v", err)
	}

	// Verify old_file.txt in dst is gone and new_file.txt exists
	if _, err := os.Stat(filepath.Join(dst, "old_file.txt")); !os.IsNotExist(err) {
		t.Errorf("expected old file in destination to be removed")
	}
	newBytes, err := os.ReadFile(filepath.Join(dst, "new_file.txt"))
	if err != nil {
		t.Fatalf("failed to read new_file.txt in dst: %v", err)
	}
	if string(newBytes) != "new content" {
		t.Errorf("unexpected content: %s", string(newBytes))
	}
}

func TestMoveDirectory_SourceErrors(t *testing.T) {
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "does_not_exist")
	dst := filepath.Join(tempDir, "dst")

	if err := MoveDirectory(nonExistent, dst); err == nil {
		t.Errorf("expected error when moving non-existent source, got nil")
	}

	fileSrc := filepath.Join(tempDir, "regular_file.txt")
	if err := os.WriteFile(fileSrc, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := MoveDirectory(fileSrc, dst); err == nil {
		t.Errorf("expected error when source is a file, got nil")
	}
}

func TestMoveFile_Success(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "src", "file.zip")
	dst := filepath.Join(tempDir, "dst", "sub", "file.zip")

	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.WriteFile(src, []byte("archive contents"), 0644); err != nil {
		t.Fatalf("failed to write src file: %v", err)
	}

	if err := MoveFile(src, dst); err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	// Verify src is gone
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("expected src to be removed, got err: %v", err)
	}

	// Verify dst content
	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read dst file: %v", err)
	}
	if string(content) != "archive contents" {
		t.Errorf("got dst content %q, want %q", string(content), "archive contents")
	}
}

func TestMoveFile_TargetExists(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "src.pdf")
	dst := filepath.Join(tempDir, "dst.pdf")

	if err := os.WriteFile(src, []byte("new pdf data"), 0644); err != nil {
		t.Fatalf("failed to write src: %v", err)
	}
	if err := os.WriteFile(dst, []byte("old pdf data"), 0644); err != nil {
		t.Fatalf("failed to write dst: %v", err)
	}

	if err := MoveFile(src, dst); err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read dst: %v", err)
	}
	if string(content) != "new pdf data" {
		t.Errorf("got dst content %q, want %q", string(content), "new pdf data")
	}
}

func TestMoveFile_Errors(t *testing.T) {
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "missing.txt")
	dst := filepath.Join(tempDir, "dst.txt")

	if err := MoveFile(nonExistent, dst); err == nil {
		t.Errorf("expected error for non-existent src, got nil")
	}

	dirSrc := filepath.Join(tempDir, "folder")
	if err := os.MkdirAll(dirSrc, 0755); err != nil {
		t.Fatalf("failed to create folder: %v", err)
	}
	if err := MoveFile(dirSrc, dst); err == nil {
		t.Errorf("expected error when src is a directory, got nil")
	}
}
