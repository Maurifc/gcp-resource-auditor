package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/maurifc/gcp-resource-auditor/internal/handlers"
)

var wg sync.WaitGroup

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
			handlers.ExportIdleExternalIPsToCSV(ctx, projectID, splitIntoDirectory, &wg)
		}
	case "instances":
		for _, projectID := range projectIDs {
			wg.Add(1)
			handlers.ExportLongTermTerminatedInstancesToCSV(ctx, projectID, 90, splitIntoDirectory, &wg)
		}
	case "firewall":
		for _, projectID := range projectIDs {
			wg.Add(1)
			handlers.ExportPermissiveRulesToCSV(ctx, projectID, splitIntoDirectory, &wg)
		}
	case "all":
		for _, projectID := range projectIDs {
			wg.Add(3)
			go handlers.ExportIdleExternalIPsToCSV(ctx, projectID, splitIntoDirectory, &wg)
			go handlers.ExportLongTermTerminatedInstancesToCSV(ctx, projectID, 90, splitIntoDirectory, &wg)
			go handlers.ExportPermissiveRulesToCSV(ctx, projectID, splitIntoDirectory, &wg)
		}
	default:
		fmt.Printf("Command '%s' not found\n", command)
		printUsage()
		os.Exit(1)
	}

	wg.Wait()
}
