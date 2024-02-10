package utils

import "os"

func EnsureTmpDirectory() error {
	tmpDir := "./tmp"
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		return os.Mkdir(tmpDir, 0755)
	}
	return nil
}
