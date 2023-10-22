package fileutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrPathNoDirectory = errors.New("invalid path, path needs a specific directory")
var ErrPathDoesNotExist = errors.New("path does not exist, no such path or directory")

// An ErrorIs function returns true if an error satisfies a particular condition.
type ErrorIs func(err error) bool

// Ignore any errors that satisfy the supplied ErrorIs function by returning
// nil. Errors that do not satisfy the supplied function are returned unmodified.
func Ignore(is ErrorIs, err error) error {
	if is(err) {
		return nil
	}
	return err
}

// ValidatePath validates if the filepath exists in the filesystem and if so
// if the filepath is a diretory
func ValidatePath(path string) (bool, error) {
	dir, path := filepath.Split(path)
	if dir == "" {
		return false, ErrPathNoDirectory
	}
	f, err := os.Stat(path)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return false, ErrPathDoesNotExist
		}
		return false, err
	}
	return f.IsDir(), nil
}

/*



type PathError = fs.PathError

&PathError{Op: "stat", Path: name, Err: err}


*/
