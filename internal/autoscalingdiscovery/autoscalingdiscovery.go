package autoscalingdiscovery

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

// GroupFilterFunc functions can be used to filter the list of autoscaling
// groups discovered by FilterGroups. Note that filtering is performed on the
// client side.
type GroupFilterFunc func(*autoscaling.Group) bool

// TagMatchFilter is a helper for filtering discovered autoscaling groups
// by any tag key+value pair.
func TagMatchFilter(key, value string) GroupFilterFunc {
	return func(group *autoscaling.Group) bool {
		for _, tag := range group.Tags {
			if *tag.Key == key && *tag.Value == value {
				return true
			}
		}
		return false
	}
}

// FilterGroups discovers autoscaling groups and applies client-side filtering
// to the results if desired. Filters are applied with AND logic. If OR logic
// is required, it can be implemented in a filter function.
//
// If any errors are encountered, a nil slice and the error are returned.
func FilterGroups(client *autoscaling.AutoScaling, filterfuncs []GroupFilterFunc) ([]*autoscaling.Group, error) {
	groups := []*autoscaling.Group{}
	params := &autoscaling.DescribeAutoScalingGroupsInput{}
	token := ""
	for {
		resp, err := client.DescribeAutoScalingGroups(params)
		if err != nil {
			return nil, err
		}
		for _, group := range resp.AutoScalingGroups {
			keep := true
			for _, filterfunc := range filterfuncs {
				if filterfunc == nil {
					panic("nil filter function passed to autoscalingdiscovery.FilterGroups")
				}
				keep = keep && filterfunc(group)
			}
			if keep {
				groups = append(groups, group)
			}
		}
		if resp.NextToken == nil {
			break
		}
		params.SetNextToken(token)
	}
	return groups, nil
}
