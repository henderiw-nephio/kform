package fsys

import (
	"errors"
	"path/filepath"
)

var ErrPathNoDirectory = errors.New("invalid path, path needs a specific directory")

// ValidateDirPath validates if the filepath is a diretory
func ValidateDirPath(path string) error {
	dir, _ := filepath.Split(path)
	if dir == "" {
		return ErrPathNoDirectory
	}
	return nil
}
