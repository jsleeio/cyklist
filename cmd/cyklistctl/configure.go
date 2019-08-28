package main

import (
	"flag"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	TagEnable    = "cyklist.jslee.io/enable"
	TagPhaseName = "cyklist.jslee.io/phase"
)

type config struct {
	Region       *string
	Session      *session.Session
	EC2          *ec2.EC2
	AutoScaling  *autoscaling.AutoScaling
	Phase        *string
	MaxInstances *int
	KubectlPath  *string
	ControlTag   *string
	EnableTag    *string
	ListOnly     *bool
}

func configureFromFlags() *config {
	config := &config{
		EnableTag:    flag.String("enable-tag", TagEnable, "AWS EC2 instance tag name to enable/disable node lifecycle processing"),
		ControlTag:   flag.String("control-tag", TagPhaseName, "AWS EC2 instance tag name to use for phase control"),
		Region:       flag.String("region", "", "AWS region to operate in"),
		Phase:        flag.String("phase", "", "Lifecycle phase to perform"),
		MaxInstances: flag.Int("max-instances", 1, "Limit number of suitably-tagged instances to operate upon"),
		KubectlPath:  flag.String("kubectl-path", "kubectl", "path to 'kubectl' executable"),
		ListOnly:     flag.Bool("list-only", false, "just list which instances would be affected"),
	}
	flag.Parse()
	config.Session = session.Must(session.NewSession(&aws.Config{
		Region:     aws.String(*config.Region),
		LogLevel:   aws.LogLevel(aws.LogOff),
		MaxRetries: aws.Int(20),
	}))
	config.EC2 = ec2.New(config.Session)
	config.AutoScaling = autoscaling.New(config.Session)
	return config
}
