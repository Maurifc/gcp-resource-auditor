package compute

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	computealt "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

type ComputeInstanceList []*computepb.Instance

func (list *ComputeInstanceList) Status(instanceStatus string) (ComputeInstanceList, error) {
	var filteredInstances ComputeInstanceList

	if list == nil || len(*list) == 0 {
		return nil, fmt.Errorf("cannot filter an empty list")
	}

	for _, instance := range *list {
		if *instance.Status == instanceStatus {
			filteredInstances = append(filteredInstances, instance)
		}
	}

	return filteredInstances, nil
}

func (list *ComputeInstanceList) FilterInstancesStoppedBefore(days int) (ComputeInstanceList, error) {
	var filteredInstances ComputeInstanceList

	if list == nil || len(*list) == 0 {
		return nil, fmt.Errorf("cannot filter an empty list")
	}

	for _, instance := range *list {
		lastStopTime, _ := time.Parse(time.RFC3339, *instance.LastStopTimestamp)
		daysDuration := time.Duration(days * 24 * int(time.Hour))
		if time.Since(lastStopTime) > daysDuration {
			filteredInstances = append(filteredInstances, instance)
		}
	}

	return filteredInstances, nil
}

func GetSummary(instance *computepb.Instance) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Name: %s\n", *instance.Name))
	sb.WriteString(fmt.Sprintf("OS: %s\n", getOs(instance)))
	sb.WriteString(fmt.Sprintf("Status: %s\n", *instance.Status))
	sb.WriteString(fmt.Sprintf("Machine Type: %s\n", getMachineType(*instance.MachineType)))
	sb.WriteString("Disks:\n")

	for _, disk := range instance.Disks {
		sb.WriteString(fmt.Sprintf("  %s: %dGB\n", *disk.DeviceName, *disk.DiskSizeGb))
	}

	sb.WriteString(fmt.Sprintf("Stop date: %s\nLast start date: %s\n", formatDate(instance.LastStopTimestamp), formatDate(instance.LastStartTimestamp)))

	return sb.String()
}

func ListAllInstances(ctx context.Context, projectID string) (ComputeInstanceList, error) {
	instancesClient, err := computealt.NewInstancesRESTClient(ctx)

	if err != nil {
		return nil, fmt.Errorf("NewInstancesRESTClient: %w", err)
	}
	defer instancesClient.Close()

	req := &computepb.AggregatedListInstancesRequest{
		Project:    projectID,
		MaxResults: proto.Uint32(3),
	}

	it := instancesClient.AggregatedList(ctx, req)
	var instances []*computepb.Instance
	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		instances = append(instances, pair.Value.Instances...)
	}

	return instances, nil
}

func formatDate(timeStamp *string) string {
	if timeStamp == nil {
		return "N/A"
	}

	parsed, err := time.Parse(time.RFC3339, *timeStamp)

	if err != nil {
		return ""
	}

	return parsed.Format(time.RFC822)
}

func getMachineType(machineTypeUrl string) string {
	re := regexp.MustCompile(`.*/machineTypes/(.*)$`)
	match := re.FindStringSubmatch(machineTypeUrl)
	return match[1]
}

func getOs(instance *computepb.Instance) string {
	return strings.Split(instance.Disks[0].Licenses[0], "licenses/")[1]
}
