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

func ExportIdleExternalIPsToCSV(ctx context.Context, projectID string, split bool, wg *sync.WaitGroup) {
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
	destinationFilePath := utils.GetDestinationFilePath(prefix, configs.IDLE_EXTERNAL_IPS_FILE)
	header := append(compute.GetStructFields(compute.IpAddressSummary{}), "ProjectID")
	err := export.ExportToCSV(header, records, destinationFilePath)

	if err != nil {
		log.Fatalf("Failure when exporting to CSV file: %s", err)
		return
	}

	fmt.Println("Records saved to", destinationFilePath)
}
