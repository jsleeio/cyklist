package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	// "github.com/jsleeio/cyklist/internal/autoscalingdiscovery"
	"github.com/jsleeio/cyklist/internal/ec2discovery"
	"github.com/jsleeio/cyklist/internal/ec2helpers"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	// AutoScalingGroupTag is derived from Amazon documentation. This tag is added to
	// new instances in an autoscaling group, and remove from those instances if they
	// are detached from that autoscaling group.
	AutoScalingGroupTag = "aws:autoscaling:groupName"
)

type config struct {
	Region      *string
	Session     *session.Session
	EC2         *ec2.EC2
	AutoScaling *autoscaling.AutoScaling
	Phase       *string
	NDetach     *int
}

func configureFromFlags() *config {
	config := &config{
		Region:  flag.String("region", "", "AWS region to operate in"),
		Phase:   flag.String("phase", "", "Lifecycle phase to perform"),
		NDetach: flag.Int("n-detach", 1, "Detach phase: number of instances to detach from their autoscaling groups"),
	}
	flag.Parse()
	config.Session = session.Must(session.NewSession(&aws.Config{
		Region:     aws.String(*config.Region),
		LogLevel:   aws.LogLevel(aws.LogOff),
		MaxRetries: 20,
	}))
	config.EC2 = ec2.New(config.Session)
	config.AutoScaling = autoscaling.New(config.Session)
	return config
}

func tagForPhase(client *ec2.EC2, phase string, instances []*string) error {
	if n := len(instances); n > 0 {
		createtags := &ec2.CreateTagsInput{
			Tags: []*ec2.Tag{
				&ec2.Tag{Key: aws.String(TagPhaseName), Value: aws.String(TagPhaseValueDetach)},
			},
			Resources: []*string{},
		}
		if n > 100 {
			n = 100
		}
		for _, instance := range instances[0:n] {
			createtags.Resources = append(createtags.Resources, instance.InstanceId)
		}
		_, err := ctx.EC2.CreateTags(createtags)
		if err != nil {
			log.Fatalf("unable to tag EC2 instances: %v", err)
		}
	}
}

func detachInstances(client *autoscaling.AutoScaling, instances []*ec2.Instance) error {
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
			return fmt.Errorf("instance %s should have tag %s, but doesn't",
				*instance.InstanceId, AutoScalingGroupTag)
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
		_, err := client.DetachInstances(detach)
		if err != nil {
			return err
		}
		log.Printf("successfully detached %d instances from group %s", len(instances), group)
	}
	return nil
}

func listInstances(instances []*ec2.Instance) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "GROUP\tINSTANCE\tIMAGE\tPHASE\tLAUNCHED\n")
	for _, instance := range instances {
		tags := ec2helpers.MapTags(instance.Tags)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			tags.GetWithDefault(AutoScalingGroupTag, "-"),
			*instance.InstanceId,
			*instance.ImageId,
			tags.GetWithDefault("amiupdate.siteminder.io/phase", "-"),
			instance.LaunchTime.UTC().Format(time.RFC3339))
	}
	w.Flush()
}

func niy(x string) {
	log.Fatalf("not implemented yet: %s", x)
}

func main() {
	cfg := configureFromFlags()
	instances, err := ec2discovery.FilterInstances(
		cfg.EC2,
		[]*ec2.Filter{
			ec2discovery.TagFilter("amiupdate.siteminder.io/enabled", "yes"),
			ec2discovery.TagFilter("amiupdate.siteminder.io/phase", *cfg.Phase),
		},
		nil)
	if err != nil {
		log.Fatalf("unable to describe instances: %v", err)
	}
	n := len(instances)
	switch *cfg.Phase {
	case "detach":
		if *cfg.NDetach <= 0 {
			log.Fatal("--n-detach argument must be greater than zero")
		}
		if n == 0 {
			log.Print("no matching instances, nothing to do")
			os.Exit(0)
		}
		if *cfg.NDetach < n {
			n = *cfg.NDetach
		}
		if err := detachInstances(cfg.AutoScaling, instances[0:n]); err != nil {
			log.Fatal("error detaching instances: %v", err)
		}
	default:
		niy(*cfg.Phase)
	}
}
