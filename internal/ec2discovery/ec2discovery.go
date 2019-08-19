package ec2discovery

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sort"
	"time"
)

// InstanceFilterFunc functions are used to filter sets of EC2 instances
type InstanceFilterFunc func(*ec2.Instance) bool

// TagFilter creates an EC2 tag filter object from a tag key and value
func TagFilter(key, value string) *ec2.Filter {
	return &ec2.Filter{
		Name:   aws.String("tag:" + key),
		Values: []*string{aws.String(value)},
	}
}

// NotImageId creates a FilterFunc that only passes instances that do
// not reference the given AMI ID
func NotImageId(ami string) InstanceFilterFunc {
	return func(i *ec2.Instance) bool {
		return *i.ImageId != ami
	}
}

// AgeAtLeast creates a FilterFunc that only passes instances older
// than a specified time.Duration
func AgeAtLeast(d time.Duration) InstanceFilterFunc {
	return func(i *ec2.Instance) bool {
		return time.Now().UTC().Sub(i.LaunchTime.UTC()) >= d
	}
}

// FilterInstances requests instance details from the EC2 API, applying EC2
// API-level filtering and optionally with one or more filter functions.
// Returns a slice of instance IDs (keys) and corresponding instance details
// (values). Returned instances will be sorted by launch time, with oldest
// instances at the left.
//
// If any errors are encountered, a nil slice and the error are returned.
func FilterInstances(client *ec2.EC2, filters []*ec2.Filter, filterfuncs []InstanceFilterFunc) ([]*ec2.Instance, error) {
	instances := []*ec2.Instance{}
	params := &ec2.DescribeInstancesInput{Filters: filters}
	token := ""
	for {
		resp, err := client.DescribeInstances(params)
		if err != nil {
			return nil, err
		}
		for _, reservation := range resp.Reservations {
			for _, instance := range reservation.Instances {
				keep := true
				for _, filterfunc := range filterfuncs {
					if filterfunc == nil {
						panic("nil filter function passed to ec2discovery.FilterInstances")
					}
					keep = keep && filterfunc(instance)
				}
				if keep {
					instances = append(instances, instance)
				}
			}
		}
		if resp.NextToken == nil {
			break
		}
		params.SetNextToken(token)
	}
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].LaunchTime.Before(*instances[j].LaunchTime)
	})
	return instances, nil
}
