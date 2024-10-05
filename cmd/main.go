package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/maurifc/gcp-resource-auditor/internal/compute"
	"github.com/maurifc/gcp-resource-auditor/internal/export"
	"github.com/maurifc/gcp-resource-auditor/internal/utils"
)

var wg sync.WaitGroup

const TERMINATED_COMPUTE_INSTANCES_FILE = "compute_instances_terminated.csv"
const IDLE_EXTERNAL_IPS_FILE = "idle_external_ips.csv"
const FIREWALL_RULES_FILE = "firewall_permissive_rules.csv"

func exportLongTermTerminatedInstancesToCSV(ctx context.Context, projectID string, daysThreshold int, split bool) {
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
	destinationFilePath := utils.GetDestinationFilePath(prefix, TERMINATED_COMPUTE_INSTANCES_FILE)
	header := append(compute.GetStructFields(compute.ComputeInstanceSummary{}), "ProjectID")
	err = export.ExportToCSV(header, records, destinationFilePath)
	if err != nil {
		log.Fatalf("Failure when exporting to CSV file", err)
	}

	fmt.Println("Records saved to", destinationFilePath)
}

func exportIdleExternalIPsToCSV(ctx context.Context, projectID string, split bool) {
	defer wg.Done()
	addresses, _ := compute.ListIpAddresses(ctx, projectID)
	externalIPs, _ := addresses.AddressType("EXTERNAL")
	idleExternalIPs, _ := externalIPs.Status("RESERVED")

	records := make([][]string, len(idleExternalIPs))
	fmt.Printf("Found %d reserved IPs that are not in use\n", len(idleExternalIPs))
	for i, ip := range idleExternalIPs {
		records[i] = append((*compute.GetIpSummary(ip)).ConvertToStringSlice(), projectID)
	}

	prefix := ""
	if split {
		prefix = projectID
	}
	destinationFilePath := utils.GetDestinationFilePath(prefix, IDLE_EXTERNAL_IPS_FILE)
	header := append(compute.GetStructFields(compute.IpAddressSummary{}), "ProjectID")
	err := export.ExportToCSV(header, records, destinationFilePath)

	if err != nil {
		log.Fatalf("Failure when exporting to CSV file", err)
	}

	fmt.Println("Records saved to", destinationFilePath)
}

func exportPermissiveRulesToCSV(ctx context.Context, projectID string, split bool) {
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

	prefix := ""
	if split {
		prefix = projectID
	}
	destinationFilePath := utils.GetDestinationFilePath(prefix, FIREWALL_RULES_FILE)
	header := append(compute.GetStructFields(compute.FirewallRuleSummary{}), "ProjectID")
	err := export.ExportToCSV(header, records, destinationFilePath)

	if err != nil {
		log.Fatalf("Failure when exporting to CSV file: ", err)
	}

	fmt.Println("Records saved to", destinationFilePath)
}

func getProjectIDsFromStdin() ([]string, error) {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("error reading stdin: %v", err)
	}

	var projectIDs []string
	for _, id := range strings.Split(string(input), ",") {
		if trimmedID := strings.TrimSpace(id); trimmedID != "" {
			projectIDs = append(projectIDs, trimmedID)
		}
	}

	return projectIDs, nil
}

func printUsage() {
	fmt.Println("Usage: go run cmd/main.go <command> [project-id1,project-id2,...] [--split]")
	fmt.Println("  or:  go run cmd/main.go <command> -")
	fmt.Println("Commands: ips, instances, firewall, all")
	fmt.Println("Use '-' to read project IDs from stdin")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/main.go firewall project-id1,project-id2")
	fmt.Println("  echo 'project-id1,project-id2' | go run cmd/main.go instances -")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	splitIntoDirectory := strings.Contains(strings.Join(os.Args, ","), "--split")
	var projectIDs []string
	var err error
	if os.Args[2] != "-" {
		projectIDs = strings.Split(os.Args[2], ",")
	} else {
		projectIDs, err = getProjectIDsFromStdin()
		if err != nil {
			log.Fatalf("Could not read Project IDs from STDIN")
			os.Exit(1)
		}
	}

	ctx := context.Background()

	switch command {
	case "ips":
		for _, projectID := range projectIDs {
			wg.Add(1)
			exportIdleExternalIPsToCSV(ctx, projectID, splitIntoDirectory)
		}
	case "instances":
		for _, projectID := range projectIDs {
			wg.Add(1)
			exportLongTermTerminatedInstancesToCSV(ctx, projectID, 90, splitIntoDirectory)
		}
	case "firewall":
		for _, projectID := range projectIDs {
			wg.Add(1)
			exportPermissiveRulesToCSV(ctx, projectID, splitIntoDirectory)
		}
	case "all":
		for _, projectID := range projectIDs {
			wg.Add(3)
			go exportIdleExternalIPsToCSV(ctx, projectID, splitIntoDirectory)
			go exportLongTermTerminatedInstancesToCSV(ctx, projectID, 90, splitIntoDirectory)
			go exportPermissiveRulesToCSV(ctx, projectID, splitIntoDirectory)
		}
	default:
		fmt.Printf("Command '%s' not found\n", command)
		printUsage()
		os.Exit(1)
	}

	wg.Wait()
}
