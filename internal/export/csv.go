package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/maurifc/gcp-resource-auditor/internal/utils"
)

func ExportToCSV(header []string, resources [][]string, destinationPath string) error {
	isAppendMode := false
	if len(resources) == 0 {
		return fmt.Errorf("no resources to export")
	}

	if _, err := os.Stat(destinationPath); err == nil {
		isAppendMode = true // file exists, then we want to append to that file
	}

	if !isAppendMode {
		utils.CreateDestinationDir(destinationPath)
	}

	file, err := os.OpenFile(destinationPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error while creating destination file: %s", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write header if it's a new file
	if !isAppendMode {
		writer.Write(header)
	}

	for _, resource := range resources {
		writer.Write(resource)
	}

	return nil
}
