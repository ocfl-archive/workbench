package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"emperror.dev/errors"
)

// EnsureTargetDirectory verifies that dir exists, or creates it if only the last path element is missing.
func EnsureTargetDirectory(dir string) error {
	fi, err := os.Stat(dir)
	if err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("target path %s exists and is not a directory", dir)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return errors.Wrapf(err, "failed to stat directory %s", dir)
	}

	parent := filepath.Dir(filepath.Clean(dir))
	pFi, err := os.Stat(parent)
	if err != nil || !pFi.IsDir() {
		return fmt.Errorf("cannot create target directory %s: parent directory %s does not exist", dir, parent)
	}

	return os.Mkdir(dir, 0755)
}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
