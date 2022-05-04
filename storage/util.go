package storage

import (
	"errors"
	"os"
)

// FileExists returns true if file exists, false if not found
func FileExists(filePath string) (bool, error) {
	logger.Printf("FileExists %s", filePath)
	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
