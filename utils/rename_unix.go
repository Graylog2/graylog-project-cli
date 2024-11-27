//go:build !windows

package utils

import (
	"github.com/google/renameio/v2"
	"os"
)

func AtomicallyWriteFile(filename string, data []byte, perm os.FileMode) error {
	if err := renameio.WriteFile(filename, data, perm); err != nil {
		return err
	}
	return nil
}
