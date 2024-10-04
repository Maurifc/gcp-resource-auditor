package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/maurifc/gcp-resource-auditor/internal/compute"
	"github.com/maurifc/gcp-resource-auditor/internal/export"
)

var wg sync.WaitGroup

const OUTPUT_DIR = "output"
const TERMINATED_COMPUTE_INSTANCES_FILE = "compute_instances_terminated.csv"
const IDLE_EXTERNAL_IPS_FILE = "idle_external_ips.csv"
const FIREWALL_RULES_FILE = "firewall_permissive_rules.csv"

func exportLongTermTerminatedInstancesToCSV(ctx context.Context, projectId string, daysThreshold int) {
	defer wg.Done()
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
		records[i] = append(compute.GetInstanceSummary(instance).ConvertToStringSlice(), projectId)
	}

	destinationFilePath := path.Join(OUTPUT_DIR, TERMINATED_COMPUTE_INSTANCES_FILE)
	header := append(compute.GetStructFields(compute.ComputeInstanceSummary{}), "ProjectID")
	err = export.ExportToCSV(header, records, destinationFilePath)
	if err != nil {
		log.Fatalf("Failure when exporting to CSV file", err)
	}

	fmt.Println("Records saved to", destinationFilePath)
}

func exportIdleExternalIPsToCSV(ctx context.Context, projectID string) {
	defer wg.Done()
	addresses, _ := compute.ListIpAddresses(ctx, projectID)
	externalIPs, _ := addresses.AddressType("EXTERNAL")
	idleExternalIPs, _ := externalIPs.Status("RESERVED")

	records := make([][]string, len(idleExternalIPs))
	fmt.Printf("Found %d reserved IPs that are not in use\n", len(idleExternalIPs))
	for i, ip := range idleExternalIPs {
		records[i] = append((*compute.GetIpSummary(ip)).ConvertToStringSlice(), projectID)
	}

	destinationFilePath := path.Join(OUTPUT_DIR, IDLE_EXTERNAL_IPS_FILE)
	header := append(compute.GetStructFields(compute.IpAddressSummary{}), "ProjectID")
	err := export.ExportToCSV(header, records, destinationFilePath)

	if err != nil {
		log.Fatalf("Failure when exporting to CSV file", err)
	}

	fmt.Println("Records saved to", destinationFilePath)
}

func exportPermissiveRulesToCSV(ctx context.Context, projectID string) {
	defer wg.Done()
	rules, _ := compute.ListAllFirewallRules(ctx, projectID)
	filteredRules, _ := rules.FilterByStatus("enabled")
	filteredRules, _ = filteredRules.FilterByAction("allow")
	filteredRules, _ = filteredRules.FilterByDirection("ingress")
	filteredRules, _ = filteredRules.FilterBySourceRange("0.0.0.0/0")

	records := make([][]string, len(filteredRules))
	fmt.Printf("Found %d permissive firewall rules\n", len(filteredRules))

	if len(filteredRules) == 0 {
		fmt.Println("Skipping...")
		return
	}

	for i, rule := range filteredRules {
		record := append((*compute.GetFirewallRuleSummary(rule)).ConvertToStringSlice(), projectID)
		records[i] = record
	}

	destinationFilePath := path.Join(OUTPUT_DIR, FIREWALL_RULES_FILE)
	header := append(compute.GetStructFields(compute.FirewallRuleSummary{}), "ProjectID")
	err := export.ExportToCSV(header, records, destinationFilePath)

	if err != nil {
		log.Fatalf("Failure when exporting to CSV file: ", err)
	}

	fmt.Println("Records saved to", destinationFilePath)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run cmd/main.go [ ips | instances | firewall | all ] <project-id>")
		os.Exit(1)
	}

	command := os.Args[1]
	projectID := os.Args[2]

	ctx := context.Background()

	switch command {
	case "ips":
		exportIdleExternalIPsToCSV(ctx, projectID)
	case "instances":
		exportLongTermTerminatedInstancesToCSV(ctx, projectID, 90)
	case "firewall":
		exportPermissiveRulesToCSV(ctx, projectID)
	case "all":
		wg.Add(3)
		go exportIdleExternalIPsToCSV(ctx, projectID)
		go exportLongTermTerminatedInstancesToCSV(ctx, projectID, 90)
		go exportPermissiveRulesToCSV(ctx, projectID)
		wg.Wait()
	default:
		fmt.Printf("Command '%s' not found\n", command)
		os.Exit(1)
	}
}
