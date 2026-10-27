// Package util provides filesystem helper functions, directory safety checks,
// atomic move/copy routines, and TCP port allocation utilities.
package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// EnsureTargetDirectory verifies that the target directory exists.
// If it does not exist, it checks whether its parent directory exists:
//   - If the parent exists, it creates the target directory.
//   - If the parent does not exist, it returns an error.
//
// It also returns an error if the path exists but is not a directory.
func EnsureTargetDirectory(dir string) error {
	fi, err := os.Stat(dir)
	if err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("target path '%s' exists but is not a directory", dir)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("error checking target path '%s': %w", dir, err)
	}

	// Target directory does not exist, check parent
	parent := filepath.Dir(dir)
	pfi, err := os.Stat(parent)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target directory '%s' cannot be created because parent directory '%s' does not exist", dir, parent)
		}
		return fmt.Errorf("error checking parent directory '%s': %w", parent, err)
	}

	if !pfi.IsDir() {
		return fmt.Errorf("parent path '%s' is not a directory", parent)
	}

	if err := os.Mkdir(dir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory '%s': %w", dir, err)
	}

	return nil
}

// MoveDirectory moves the directory at src to dst.
// It creates the parent directory of dst if needed, removes any existing collision at dst,
// and supports cross-device moves by falling back to recursive copy and remove.
func MoveDirectory(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source directory error: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source '%s' is not a directory", src)
	}

	// Ensure parent directory of dst exists
	dstParent := filepath.Dir(dst)
	if err := os.MkdirAll(dstParent, 0755); err != nil {
		return fmt.Errorf("failed to create destination parent directory '%s': %w", dstParent, err)
	}

	// If destination already exists, remove it to prevent collision
	if _, err := os.Stat(dst); err == nil {
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("failed to remove existing destination directory '%s': %w", dst, err)
		}
	}

	// Attempt atomic rename
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fallback for cross-device moves: copy recursively then delete source
	if err := copyDir(src, dst); err != nil {
		_ = os.RemoveAll(dst)
		return fmt.Errorf("failed to copy directory from '%s' to '%s': %w", src, dst, err)
	}

	if err := os.RemoveAll(src); err != nil {
		return fmt.Errorf("failed to remove source directory '%s' after copy: %w", src, err)
	}

	return nil
}

// MoveFile moves the file at src to dst.
// It creates the parent directory of dst if needed, removes any existing collision at dst,
// and supports cross-device moves by falling back to copy and remove.
func MoveFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source file error: %w", err)
	}
	if srcInfo.IsDir() {
		return fmt.Errorf("source '%s' is a directory", src)
	}

	// Ensure parent directory of dst exists
	dstParent := filepath.Dir(dst)
	if err := os.MkdirAll(dstParent, 0755); err != nil {
		return fmt.Errorf("failed to create destination parent directory '%s': %w", dstParent, err)
	}

	// If destination already exists, remove it to prevent collision
	if _, err := os.Stat(dst); err == nil {
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("failed to remove existing destination file '%s': %w", dst, err)
		}
	}

	// Attempt atomic rename
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fallback for cross-device moves: copy file then delete source
	if err := copyFile(src, dst); err != nil {
		_ = os.Remove(dst)
		return fmt.Errorf("failed to copy file from '%s' to '%s': %w", src, dst, err)
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source file '%s' after copy: %w", src, err)
	}

	return nil
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	srcInfo, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
