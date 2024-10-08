package compute

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

type FirewallRuleList []*computepb.Firewall

type FirewallRuleSummary struct {
	Name         string
	Action       string
	AllowedPorts []string
	DeniedPorts  []string
	Disabled     bool
	Direction    string
	SourceRanges []string
}

func GetFirewallRuleSummary(rule *computepb.Firewall) *FirewallRuleSummary {
	summary := FirewallRuleSummary{
		Name:         *rule.Name,
		Action:       getRuleAction(rule),
		AllowedPorts: make([]string, 0),
		DeniedPorts:  make([]string, 0),
		Disabled:     *rule.Disabled,
		Direction:    *rule.Direction,
		SourceRanges: rule.SourceRanges,
	}

	// fillout allowed ports
	for _, allowRule := range rule.Allowed {
		summary.AllowedPorts = append(summary.AllowedPorts, allowRule.Ports...)
	}

	// fillout denied ports
	for _, denyRule := range rule.Denied {
		summary.DeniedPorts = append(summary.DeniedPorts, denyRule.Ports...)
	}

	return &summary
}

func (ruleSummary *FirewallRuleSummary) ConvertToStringSlice() []string {
	return []string{
		ruleSummary.Name,
		ruleSummary.Action,
		strings.Join(ruleSummary.AllowedPorts, ";"),
		strings.Join(ruleSummary.DeniedPorts, ";"),
		strconv.FormatBool(ruleSummary.Disabled),
		ruleSummary.Direction,
		strings.Join(ruleSummary.SourceRanges, ";"),
	}
}

func (ruleList *FirewallRuleList) FilterByDirection(direction string) (FirewallRuleList, error) {
	var filteredList FirewallRuleList

	if ruleList == nil || len(*ruleList) == 0 {
		return nil, fmt.Errorf("cannot filter an empty list")
	}

	for _, rule := range *ruleList {
		if strings.ToUpper(direction) == *rule.Direction {
			filteredList = append(filteredList, rule)
		}
	}

	return filteredList, nil
}

func (ruleList *FirewallRuleList) FilterByStatus(status string) (FirewallRuleList, error) {
	var filteredList FirewallRuleList

	if ruleList == nil || len(*ruleList) == 0 {
		return nil, fmt.Errorf("cannot filter an empty list")
	}

	for _, rule := range *ruleList {
		if strings.ToLower(status) == "enabled" && *rule.Disabled == false {
			filteredList = append(filteredList, rule)
		} else if strings.ToLower(status) == "disabled" && *rule.Disabled == true {
			filteredList = append(filteredList, rule)
		}
	}

	return filteredList, nil
}

func (ruleList *FirewallRuleList) FilterByPort(port string) (FirewallRuleList, error) {
	var filteredList FirewallRuleList

	if ruleList == nil || len(*ruleList) == 0 {
		return nil, fmt.Errorf("cannot filter an empty list")
	}

	for _, rule := range *ruleList {
		for _, allowedRule := range rule.Allowed {
			for _, portRange := range allowedRule.Ports {
				if strings.Contains(portRange, port) {
					filteredList = append(filteredList, rule)
				}
			}
		}
	}

	return filteredList, nil
}

func (ruleList *FirewallRuleList) FilterBySourceRange(ipRange string) (FirewallRuleList, error) {
	var filteredList FirewallRuleList

	if ruleList == nil || len(*ruleList) == 0 {
		return nil, fmt.Errorf("cannot filter an empty list")
	}

	for _, rule := range *ruleList {
		for _, ips := range rule.SourceRanges {
			if strings.Contains(ips, ipRange) {
				filteredList = append(filteredList, rule)
			}
		}
	}

	return filteredList, nil
}

func (ruleList *FirewallRuleList) FilterByAction(action string) (FirewallRuleList, error) {
	var filteredList FirewallRuleList

	if ruleList == nil || len(*ruleList) == 0 {
		return nil, fmt.Errorf("cannot filter an empty list")
	}

	for _, rule := range *ruleList {
		if getRuleAction(rule) == action {
			filteredList = append(filteredList, rule)
		}
	}

	return filteredList, nil
}

func ListAllFirewallRules(ctx context.Context, projectID string) (FirewallRuleList, error) {
	var rules FirewallRuleList

	firewallsClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewInstancesRESTClient: %w", err)
	}
	defer firewallsClient.Close()

	req := &computepb.ListFirewallsRequest{
		Project: projectID,
	}

	it := firewallsClient.List(ctx, req)
	for {
		firewallRule, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		rules = append(rules, firewallRule)
	}

	return rules, nil
}

func getRuleAction(rule *computepb.Firewall) string {
	if len(rule.Allowed) > 0 {
		return "allow"
	}

	return "deny"
}
