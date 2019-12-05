package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/elbv2"

	"github.com/sensu/sensu-aws/awsclient"

	"github.com/sensu/sensu-aws/aws_session"

	"github.com/aws/aws-sdk-go/aws/session"
)

/*
#
# check-alb-target-group-health
#
# DESCRIPTION:
#   This plugin checks the health of Application Load Balancer target groups
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   Check all target groups in a region
#   ./check-alb-target-group-health --aws_region=us-east-1
#
#   Check a single target group in a region
#   ./check-alb-target-group-health --aws_region=us-east-1 --target_groups=target-group-1
#
#   Check multiple target groups in a region
#   ./check-alb-target-group-health --aws_region=us-east-1 --target_groups=target-group-a,target-group-b
#
# LICENSE:
#   TODO
*/

var (
	targetGroups string
	awsRegion    string
	critical     bool
	elbV2lient   *elbv2.ELBV2
	awsSession   *session.Session
)

func checkHealth() {
	var success bool
	awsSession = aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, elbV2lient = awsclient.GetElbV2Client(awsSession)
	if !success {
		return
	}
	targetGroups, err := getTargerGroups()
	if err != nil || targetGroups == nil {
		return
	}
	getUnhealthyTargetGroupCount(targetGroups)
}

func getTargerGroups() ([]*elbv2.TargetGroup, error) {
	targets := strings.Split(targetGroups, ",")
	input := &elbv2.DescribeTargetGroupsInput{}
	if targets != nil && len(targets) > 0 {
		input.Names = aws.StringSlice(targets)
	}
	output, err := elbV2lient.DescribeTargetGroups(input)
	if err != nil {
		fmt.Println("Error while calling DescribeTargetGroups AWS API,", err.(awserr.Error).Message())
		return nil, err
	}
	if !(output != nil && output.TargetGroups != nil && len(output.TargetGroups) > 0) {
		return nil, nil
	}
	return output.TargetGroups, nil
}

func getUnhealthyTargetGroupCount(targetGroups []*elbv2.TargetGroup) {
	unhealthyTargetGroups := make(map[string]int)
	unhealthyTargets := make(map[string][]string)
	for _, targetGroup := range targetGroups {
		healthInput := &elbv2.DescribeTargetHealthInput{}
		healthInput.TargetGroupArn = targetGroup.TargetGroupArn
		healthOutput, err := elbV2lient.DescribeTargetHealth(healthInput)
		if err != nil {
			fmt.Println("Error while calling DescribeTargetHealth AWS API,", err.(awserr.Error).Message())
			return
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
		fmt.Println("OK : All ALB target groups are healthy")
		return
	}
	if critical {
		fmt.Println("CRITICAL:")
	} else {
		fmt.Println("WARNING:")
	}
	for targetGroup, count := range unhealthyTargetGroups {
		fmt.Println(fmt.Sprintf("Target Group Name : '%s' having %d unhealthy targets - %v", targetGroup, count, unhealthyTargets[targetGroup]))
	}
}

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return fmt.Errorf("invalid argument(s) received")
	}
	checkHealth()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-alb-target-group-health",
		Short: "The Sensu Go Aws ALB handler for health management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&targetGroups, "target_groups", "", "The ALB target group(s) to check. Separate multiple target groups with commas")
	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	cmd.Flags().BoolVar(&critical, "critical", false, "Critical instead of warn when unhealthy targets are found")

	return cmd
}
