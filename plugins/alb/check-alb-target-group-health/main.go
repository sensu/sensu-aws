package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

var (
	targetGroups []string
	awsRegion    string
	critical     bool
	awsSession   *session.Session

	config = &sensu.PluginConfig{
		Name:     "check-alb-target-group-health",
		Short:    "The Sensu Go Aws ALB check for health management",
		Timeout:  10,
		Keyspace: "sensu.io/plugins/sensu-aws/check-alb-target-group-health",
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:     "target-groups",
			Env:      "TARGET_GROUPS",
			Argument: "target-groups",
			Usage:    "The ALB target group(s) to check",
			Value:    &targetGroups,
		},
		{
			Path:     "aws-region",
			Env:      "AWS_REGION",
			Argument: "aws-region",
			Usage:    "AWS Region",
			Default:  "us-east-1",
			Value:    &awsRegion,
		},
		{
			Path:     "critical",
			Env:      "CRITICAL",
			Argument: "critical",
			Default:  false,
			Usage:    "Critical instead of warn when unhealthy targets are found",
			Value:    &critical,
		},
	}
)

// ELBClient represents the external dependencies of checkHealth()
type ELBClient interface {
	DescribeTargetGroups(*elbv2.DescribeTargetGroupsInput) (*elbv2.DescribeTargetGroupsOutput, error)
	DescribeTargetHealth(*elbv2.DescribeTargetHealthInput) (*elbv2.DescribeTargetHealthOutput, error)
}

func checkHealth(client ELBClient, targets []string, critical bool) (int, error) {
	targetGroups, err := getTargetGroups(client, targets)
	if err != nil {
		return 2, err
	}
	unhealthyTargetGroups := make(map[string]int)
	unhealthyTargets := make(map[string][]string)
	for _, targetGroup := range targetGroups {
		healthInput := &elbv2.DescribeTargetHealthInput{}
		healthInput.TargetGroupArn = targetGroup.TargetGroupArn
		healthOutput, err := client.DescribeTargetHealth(healthInput)
		if err != nil {
			return 2, err
		}
		if !(healthOutput != nil && healthOutput.TargetHealthDescriptions != nil && len(healthOutput.TargetHealthDescriptions) > 0) {
			continue
		}
		for _, target := range healthOutput.TargetHealthDescriptions {
			if *target.TargetHealth.State == "unhealthy" {
				unhealthyTargetGroups[*targetGroup.TargetGroupName] = unhealthyTargetGroups[*targetGroup.TargetGroupName] + 1
				unhealthyTargets[*targetGroup.TargetGroupName] = append(unhealthyTargets[*targetGroup.TargetGroupName], *target.Target.Id)
			}
		}
	}
	if len(unhealthyTargetGroups) <= 0 {
		return 0, nil
	}
	status := 1
	if critical {
		status = 2
	}
	for targetGroup, count := range unhealthyTargetGroups {
		log.Println(fmt.Sprintf("Target group '%s' has %d unhealthy members - %v", targetGroup, count, unhealthyTargets[targetGroup]))
	}
	return status, errors.New("one or more target groups is unhealthy")
}

func getTargetGroups(client ELBClient, targetGroups []string) ([]*elbv2.TargetGroup, error) {
	if len(targetGroups) == 0 {
		return nil, errors.New("no target groups specified")
	}
	input := &elbv2.DescribeTargetGroupsInput{
		Names: aws.StringSlice(targetGroups),
	}
	output, err := client.DescribeTargetGroups(input)
	if err != nil {
		return nil, err
	}
	return output.TargetGroups, nil
}

func main() {
	validator := func(*corev2.Event) (int, error) {
		return 0, nil
	}
	executor := func(*corev2.Event) (int, error) {
		session, err := session.NewSession(&aws.Config{
			Region: &awsRegion,
		})
		if err != nil {
			return 2, err
		}
		return checkHealth(elbv2.New(session), targetGroups, critical)
	}
	sensu.NewGoCheck(config, options, validator, executor, false).Execute()
}
