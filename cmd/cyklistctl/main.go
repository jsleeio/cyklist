package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/jsleeio/cyklist/internal/ec2discovery"
	"github.com/jsleeio/cyklist/internal/ec2helpers"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	// AutoScalingGroupTag is derived from Amazon documentation. This tag is added to
	// new instances in an autoscaling group, and remove from those instances if they
	// are detached from that autoscaling group.
	AutoScalingGroupTag = "aws:autoscaling:groupName"
)

func minint(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func tagForPhase(phase string, instances []*ec2.Instance, cfg *config) error {
	log.Printf("got %d instances for tagging", len(instances))
	n := minint(100, len(instances))
	if n < 1 {
		return nil
	}
	createtags := &ec2.CreateTagsInput{
		Tags: []*ec2.Tag{
			&ec2.Tag{Key: aws.String(*cfg.ControlTag), Value: aws.String(phase)},
		},
		Resources: []*string{},
	}
	for _, instance := range instances[0:n] {
		createtags.Resources = append(createtags.Resources, instance.InstanceId)
	}
	_, err := cfg.EC2.CreateTags(createtags)
	return err
}

func listInstances(instances []*ec2.Instance, controltag string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "GROUP\tINSTANCE\tIMAGE\tPHASE\tLAUNCHED\n")
	for _, instance := range instances {
		tags := ec2helpers.MapTags(instance.Tags)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			tags.GetWithDefault(AutoScalingGroupTag, "-"),
			*instance.InstanceId,
			*instance.ImageId,
			tags.GetWithDefault(controltag, "-"),
			instance.LaunchTime.UTC().Format(time.RFC3339))
	}
	w.Flush()
}

func executeDetach(instances []*ec2.Instance, cfg *config) error {
	// phase 1: map lists of instances to autoscaling groups
	groupToInstances := make(map[string][]*string)
	for _, instance := range instances {
		autoscalingGroup := ""
		for _, tag := range instance.Tags {
			if *tag.Key == AutoScalingGroupTag {
				autoscalingGroup = *tag.Value
			}
		}
		if autoscalingGroup == "" {
			log.Printf("instance %s should have tag %s, but doesn't. Tagging for drain to make sure it eventually gets cleaned up",
				*instance.InstanceId, AutoScalingGroupTag)
			tagForPhase("drain", []*ec2.Instance{instance}, cfg)
			continue
		}
		ii, ok := groupToInstances[autoscalingGroup]
		if ok {
			ii = append(ii, instance.InstanceId)
		} else {
			ii = []*string{instance.InstanceId}
		}
		groupToInstances[autoscalingGroup] = ii
	}
	// phase 2: detach instances from their groups
	for group, instances := range groupToInstances {
		for _, instance := range instances {
			fmt.Printf("group %s instance %s\n", group, *instance)
		}
		detach := &autoscaling.DetachInstancesInput{
			AutoScalingGroupName:           aws.String(group),
			InstanceIds:                    instances,
			ShouldDecrementDesiredCapacity: aws.Bool(false),
		}
		_, err := cfg.AutoScaling.DetachInstances(detach)
		if err != nil {
			return err
		}
		log.Printf("successfully detached %d instances from group %s", len(instances), group)
	}
	return nil
}

func executeDrain(instances []*ec2.Instance, cfg *config) error {
	for _, instance := range instances {
		log.Printf("%s: draining node", *instance.PrivateDnsName)
		draincmd := exec.Command(*cfg.KubectlPath, "drain", "--timeout=1h",
			"--ignore-daemonsets", "--delete-local-data", "--force",
			*instance.PrivateDnsName)
		if err := draincmd.Run(); err != nil {
			log.Printf("%s: error draining node: %v", *instance.PrivateDnsName, err)
			return err
		} else {
			log.Printf("%s: successfully drained (except daemonsets)", *instance.PrivateDnsName)
		}
	}
	return nil
}

func executeTerminate(instances []*ec2.Instance, cfg *config) error {
	tii := &ec2.TerminateInstancesInput{InstanceIds: []*string{}}
	for _, instance := range instances {
		tii.InstanceIds = append(tii.InstanceIds, instance.InstanceId)
	}
	_, err := cfg.EC2.TerminateInstances(tii)
	return err
}

func main() {
	cfg := configureFromFlags()
	instances, err := ec2discovery.FilterInstances(
		cfg.EC2,
		[]*ec2.Filter{
			ec2discovery.TagFilter(*cfg.EnableTag, "yes"),
			ec2discovery.TagFilter(*cfg.ControlTag, *cfg.Phase),
		},
		nil)
	if err != nil {
		log.Fatalf("unable to describe instances: %v", err)
	}
	// always pick the oldest instances first
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].LaunchTime.Unix() < instances[j].LaunchTime.Unix()
	})
	subset := instances[0:minint(*cfg.MaxInstances, len(instances))]
	if len(subset) < 1 {
		// nothing to do
		log.Printf("%s: nothing to do!", *cfg.Phase)
		os.Exit(0)
	}
	if *cfg.ListOnly {
		listInstances(subset, *cfg.ControlTag)
		os.Exit(0)
	}
	var next string
	switch *cfg.Phase {
	case "detach":
		err = executeDetach(subset, cfg)
		next = "drain"
	case "drain":
		err = executeDrain(subset, cfg)
		next = "terminate"
	case "terminate":
		err = executeTerminate(subset, cfg)
		next = ""
	default:
		log.Fatalf("unknown phase: %s", *cfg.Phase)
	}
	if err != nil {
		log.Fatalf("%s: error executing phase, not continuing: %v", *cfg.Phase, err)
	}
	if err = tagForPhase(next, subset, cfg); err != nil {
		log.Fatalf("%s: error tagging %d instances for next phase %s: %v", *cfg.Phase, len(subset), next, err)
	} else {
		if next != "" {
			log.Printf("%s: operated on %d instances and tagged them for next phase: %s", *cfg.Phase, len(subset), next)
		} else {
			log.Printf("%s: operated on %d instances", *cfg.Phase, len(subset))
		}
	}
}
