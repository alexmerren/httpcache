package httpcache

import (
	"errors"
	"os"
	"path/filepath"
)

func doesFileExist(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return true, nil
}

func createFile(filename string) (*os.File, error) {
	directory, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return os.Create(filename)
}
