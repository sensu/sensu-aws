package main

/*
#
# check-elb-health-fog
#
#
# DESCRIPTION:
#   This plugin checks the health of an Amazon Elastic Load Balancer.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#  ./check-elb-health-fog -aws_region=${you_region} --instances=${your_instance_ids} --elb_name=${your_elb_name} --verbose=true
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/awsclient"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sensu/sensu-aws/aws_session"
)

var (
	awsRegion string
	elbName   string
	instances string
	verbose   bool
	elbClient *elb.ELB
)

func checkHealth() {
	var success bool
	awsSession := aws_session.CreateAwsSession()
	success, elbClient = awsclient.GetElbClient(awsSession)
	if !success {
		return
	}

	instanceStates, err := getInstanceHealth()
	if err != nil {
		fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
		return
	}
	checkUnhealthyInstances(instanceStates)
}

func getInstanceHealth() ([]*elb.InstanceState, error) {
	instanceIdentifiers := strings.Split(instances, ",")
	input := &elb.DescribeInstanceHealthInput{}
	for _, instanceID := range instanceIdentifiers {
		input.Instances = append(input.Instances, &elb.Instance{InstanceId: &instanceID})
	}
	input.LoadBalancerName = &elbName
	output, err := elbClient.DescribeInstanceHealth(input)
	if err != nil {
		return nil, err
	}
	if !(output != nil && output.InstanceStates != nil && len(output.InstanceStates) > 0) {
		return nil, nil
	}
	return output.InstanceStates, nil
}

func checkUnhealthyInstances(instanceStates []*elb.InstanceState) {
	unhealthyInstances := make(map[string]string)
	for _, instanceState := range instanceStates {
		if *instanceState.State != "InService" {
			unhealthyInstances[*instanceState.InstanceId] = *instanceState.State
		}
	}

	if unhealthyInstances == nil || len(unhealthyInstances) <= 0 {
		fmt.Println("OK : All instances on ELB ", awsRegion, "::", elbName, "healthy!")
		return
	}

	if verbose {
		fmt.Println("CRITICAL : Unhealthy instances detected:")
		for id, state := range unhealthyInstances {
			fmt.Println(id, "::", state)
		}
	} else {
		fmt.Println("CRITICAL : Detected ", len(unhealthyInstances), "unhealthy instances")
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
		Use:   "check-elb-health-fog",
		Short: "The Sensu Go Aws Load Balancer handler for health management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "eu-west-1", "AWS Region (such as eu-west-1). If you do not specify a region, it will be detected by the server the script is run on")
	cmd.Flags().StringVar(&elbName, "elb_name", "", "The Elastic Load Balancer name of which you want to check the health")
	cmd.Flags().StringVar(&instances, "instances", "", "Comma separated list of specific instances IDs inside the ELB of which you want to check the health")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable a little bit more verbose reports about instance health")
	_ = cmd.MarkFlagRequired("elb_name")
	return cmd
}
