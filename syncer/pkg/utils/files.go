package utils

import (
	"os"
	"path/filepath"
)

// IsDirExist checks if a dir exists
func IsDirExist(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir() || os.IsExist(err)
}

// IsFileExist checks if a file exists
func IsFileExist(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir() || os.IsExist(err)
}

// OpenFile if file not exist auto create
func OpenFile(path string) (*os.File, error)  {
	if IsFileExist(path) {
		return os.Open(path)
	}

	dir := filepath.Dir(path)
	if !IsDirExist(dir) {
		err := os.MkdirAll(dir, 0440)
		if err != nil {
			return nil , err
		}
	}
	return os.Create(path)
}
