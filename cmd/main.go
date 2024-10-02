package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/maurifc/gcp-resource-auditor/internal/compute"
	"github.com/maurifc/gcp-resource-auditor/internal/export"
)

func displayLongTermTerminatedInstances(ctx context.Context, projectId string, daysThreshold int) {
	instances, err := compute.ListAllInstances(ctx, projectId)
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

	fmt.Printf(">>> Listing instances terminated for more than %d days <<<\n", daysThreshold)
	longTermTerminatedInstances, err := terminatedInstances.FilterInstancesStoppedBefore(daysThreshold)

	if err != nil {
		log.Fatalf("Failed to filter long-term terminated instances: %v", err)
	}

	if len(longTermTerminatedInstances) == 0 {
		fmt.Printf("No instances found that have been terminated for more than %d days.\n", daysThreshold)
		return
	}

	for _, instance := range longTermTerminatedInstances {
		fmt.Println(compute.GetSummary(instance))
	}
}

func exportIdleExternalIPsToCSV(ctx context.Context, projectID string) {
	addresses, _ := compute.ListIpAddresses(ctx, projectID)
	externalIPs, _ := addresses.AddressType("EXTERNAL")
	idleExternalIPs, _ := externalIPs.Status("RESERVED")

	records := make([][]string, len(idleExternalIPs))
	fmt.Printf("Found %d reserved IPs that are not in use\n", len(idleExternalIPs))
	for i, ip := range idleExternalIPs {
		records[i] = (*compute.GetIpSummary(ip)).ConvertToStringSlice()
	}

	err := export.ExportToCSV(records, "idle_external_ips.csv")

	if err != nil {
		log.Fatalf("Failure when exporting to CSV file")
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <project-id> [ ips | instances ]")
		os.Exit(1)
	}

	command := os.Args[1]
	projectID := os.Args[2]

	ctx := context.Background()
	switch command {
	case "ips":
		exportIdleExternalIPsToCSV(ctx, projectID)
	case "instances":
		displayLongTermTerminatedInstances(ctx, projectID, 90)
	default:
		fmt.Printf("Command '%s' not found\n", command)
		os.Exit(1)
	}
}
