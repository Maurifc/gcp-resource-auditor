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

func ExportPermissiveRulesToCSV(ctx context.Context, projectID string, split bool, wg *sync.WaitGroup) {
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
	destinationFilePath := utils.GetDestinationFilePath(prefix, configs.FIREWALL_RULES_FILE)
	header := append(compute.GetStructFields(compute.FirewallRuleSummary{}), "ProjectID")
	err := export.ExportToCSV(header, records, destinationFilePath)

	if err != nil {
		log.Fatalf("Failure when exporting to CSV file: %s", err)
	}

	fmt.Println("Records saved to", destinationFilePath)
}
