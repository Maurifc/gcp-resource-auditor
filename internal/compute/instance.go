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

type ComputeInstanceSummary struct {
	Name          string
	OS            string
	Status        string
	MachineType   string
	LastStartDate string
	LastStopDate  string
}

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

func GetInstanceSummary(instance *computepb.Instance) *ComputeInstanceSummary {
	return &ComputeInstanceSummary{
		Name:          *instance.Name,
		OS:            getOs(instance),
		Status:        *instance.Status,
		MachineType:   getMachineType(*instance.MachineType),
		LastStartDate: *instance.LastStartTimestamp,
		LastStopDate:  *instance.LastStopTimestamp,
	}
}

func (instanceSummary *ComputeInstanceSummary) ConvertToStringSlice() []string {
	return []string{
		instanceSummary.Name,
		instanceSummary.OS,
		instanceSummary.Status,
		instanceSummary.MachineType,
		instanceSummary.LastStartDate,
		instanceSummary.LastStopDate,
	}
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
