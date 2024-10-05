package handlers

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/maurifc/gcp-resource-auditor/internal/compute"
	"github.com/maurifc/gcp-resource-auditor/internal/configs"
	"github.com/maurifc/gcp-resource-auditor/internal/export"
	"github.com/maurifc/gcp-resource-auditor/internal/utils"
)

func ExportLongTermTerminatedInstancesToCSV(ctx context.Context, projectID string, daysThreshold int, split bool, wg *sync.WaitGroup) {
	defer wg.Done()
	instances, err := compute.ListAllInstances(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to retrieve instances: %v", err)
	}

	terminatedInstances, err := instances.Status("TERMINATED")
	if err != nil {
		log.Fatalf("Failed to filter terminated instances: %v", err)
	}

	if len(terminatedInstances) == 0 {
		fmt.Println("No terminated instances found.")
		return
	}

	longTermTerminatedInstances, err := terminatedInstances.FilterInstancesStoppedBefore(daysThreshold)

	if err != nil {
		log.Fatalf("Failed to filter long-term terminated instances: %v", err)
	}

	if len(longTermTerminatedInstances) == 0 {
		fmt.Printf("No instances found that have been terminated for more than %d days.\n", daysThreshold)
		return
	}

	records := make([][]string, len(longTermTerminatedInstances))

	fmt.Printf("Found %d long term terminated Instances\n", len(longTermTerminatedInstances))
	for i, instance := range longTermTerminatedInstances {
		records[i] = append(compute.GetInstanceSummary(instance).ConvertToStringSlice(), projectID)
	}

	prefix := ""
	if split {
		prefix = projectID
	}
	destinationFilePath := utils.GetDestinationFilePath(prefix, configs.TERMINATED_COMPUTE_INSTANCES_FILE)
	header := append(compute.GetStructFields(compute.ComputeInstanceSummary{}), "ProjectID")
	err = export.ExportToCSV(header, records, destinationFilePath)
	if err != nil {
		log.Fatalf("Failure when exporting to CSV file: %s", err)
	}

	fmt.Println("Records saved to", destinationFilePath)
}
