package compute

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

type IpAddressList []*computepb.Address

type IpAddressSummary struct {
	Name        string
	Status      string
	Address     string
	AddressType string
}

func (list *IpAddressList) AddressType(addressType string) (IpAddressList, error) {
	var filteredList IpAddressList

	if len(*list) == 0 {
		return nil, fmt.Errorf("could not filter an empty list")
	}

	for _, ip := range *list {
		if *ip.AddressType == addressType {
			filteredList = append(filteredList, ip)
		}
	}

	return filteredList, nil
}

func (list *IpAddressList) Status(addressStatus string) (IpAddressList, error) {
	var filteredList IpAddressList

	if len(*list) == 0 {
		return nil, fmt.Errorf("could not filter an empty list")
	}

	for _, ip := range *list {
		if *ip.Status == addressStatus {
			filteredList = append(filteredList, ip)
		}
	}

	return filteredList, nil
}

func ListIpAddresses(ctx context.Context, projectID string) (IpAddressList, error) {
	addressesClient, err := compute.NewAddressesRESTClient(ctx)

	if err != nil {
		return nil, fmt.Errorf("NewAddressesRESTClient: %w", err)
	}

	defer addressesClient.Close()

	req := &computepb.AggregatedListAddressesRequest{
		Project:    projectID,
		MaxResults: proto.Uint32(3),
	}

	it := addressesClient.AggregatedList(ctx, req)
	var addresses []*computepb.Address

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		addresses = append(addresses, pair.Value.Addresses...)
	}

	return addresses, nil
}

func GetIpSummary(ipAddress *computepb.Address) *IpAddressSummary {
	return &IpAddressSummary{
		Name:        *ipAddress.Name,
		Status:      *ipAddress.Status,
		Address:     *ipAddress.Address,
		AddressType: *ipAddress.AddressType,
	}
}

func (ipSummary *IpAddressSummary) ConvertToStringSlice() []string {
	return []string{
		ipSummary.Name,
		ipSummary.Status,
		ipSummary.Address,
		ipSummary.AddressType,
	}
}
