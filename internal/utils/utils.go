package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetProjectRoot() (string, error) {
	return os.Getwd()
}

// gets the project root directory name and returns it as the name for go build output binary
func GetBinaryName() string {
	cwd, err := GetProjectRoot()
	if err != nil {
		log.Fatalf("Failed to get working directory:%v\n", err)
	}
	name := filepath.Base(cwd)
	if len(name) > 0 {
		return name
	}
	return "main"
}

func CleanPath(path string) string {
	return strings.TrimSuffix(strings.TrimSpace(path), "/")
}
