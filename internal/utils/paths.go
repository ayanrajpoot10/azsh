package utils

import (
	"os"
	"path/filepath"
)

const configDir = ".azsh"

// CachePath returns the absolute path to ~/.azsh/<filename>,
// creating the ~/.azsh directory if it doesn't exist.
func CachePath(filename string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}
