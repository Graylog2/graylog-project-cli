//go:build windows

package utils

import (
	"os"
	"path/filepath"
)

func AtomicallyWriteFile(filename string, data []byte, perm os.FileMode) error {
	// The github.com/google/renameio library doesn't support Windows so we have to do a best-effort implementation.
	file, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".*")
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	file.Close()
	if err := os.Rename(file.Name(), filename); err != nil {
		return err
	}
	return nil
}
