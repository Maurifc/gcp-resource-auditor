package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/maurifc/gcp-resource-auditor/internal/utils"
)

func ExportToCSV(header []string, resources [][]string, destinationPath string) error {
	if len(resources) == 0 {
		return fmt.Errorf("no resources to export")
	}

	utils.CreateDestinationDir(destinationPath)

	file, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("error while creating destination file: %s", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write header first
	writer.Write(header)
	for _, resource := range resources {
		writer.Write(resource)
	}

	return nil
}
