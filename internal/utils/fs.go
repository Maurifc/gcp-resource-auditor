package utils

import (
	"os"
	"path"
	"path/filepath"
)

const OUTPUT_DIR = "output"

func GetDestinationFilePath(projectID string, fileName string) string {
	destinationDir := path.Join(OUTPUT_DIR, projectID)
	destinationFilePath := path.Join(destinationDir, fileName)
	return destinationFilePath
}

func CreateDestinationDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}
