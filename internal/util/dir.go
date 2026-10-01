package util

import (
	"fmt"
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
