package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/awsclient"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/elb"

	"github.com/sensu/sensu-aws/aws_session"

	"github.com/aws/aws-sdk-go/aws/session"
)

/*
#
# check-elb-nodes
#
# DESCRIPTION:
#   This plugin checks an AWS Elastic Load Balancer to ensure a minimum number
#   or percentage of nodes are InService on the ELB
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   Warning if the load balancer has 3 or fewer healthy nodes and critical if 2 or fewer
#   ./check-elb-nodes --warning=3 --critical=2 --load_balancer=#{your-load-balancer}
#
#   Warning if the load balancer has 50% or less healthy nodes and critical if 25% or less
#   ./check-elb-nodes --warning_percentage=50 --critical_percentage=25 --load_balancer=#{your-load-balancer}
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

var (
	awsregion          string
	elbName            string
	warning            int
	critical           int
	warningPercentage  float64
	criticalPercentage float64
	elbClient          *elb.ELB
)

func checkNodes() {
	var awsSession *session.Session
	var success bool
	if len(elbName) <= 0 {
		fmt.Println("Please enter a load balance name")
		return
	}
	if (critical == -1 || warning == -1) && (criticalPercentage == -1 || warningPercentage == -1) {
		fmt.Println("please enter (critical and warning non zero positive value) and/or (critical percentage and warning percentage non zero positive value)")
		return
	}
	awsSession = aws_session.CreateAwsSessionWithRegion(awsregion)
	success, elbClient = awsclient.GetElbClient(awsSession)
	if !success {
		return
	}
	instanceStates, err := getInstanceHealth()
	if err != nil || instanceStates == nil {
		return
	}
	checkInstanceHealth(instanceStates)
}

func getInstanceHealth() ([]*elb.InstanceState, error) {
	input := &elb.DescribeInstanceHealthInput{}
	input.LoadBalancerName = &elbName
	output, err := elbClient.DescribeInstanceHealth(input)
	if err != nil && err.(awserr.Error).Code() == "LoadBalancerNotFound" {
		fmt.Println(err.(awserr.Error).Message())
		return nil, err
	} else if err != nil {
		fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
		return nil, err
	}
	if !(output != nil && output.InstanceStates != nil && len(output.InstanceStates) > 0) {
		return nil, nil
	}
	return output.InstanceStates, nil
}

func checkInstanceHealth(instanceStates []*elb.InstanceState) {
	var pecentage float64
	instancesStatesCountMap := make(map[string]int)
	for _, instanceState := range instanceStates {
		instancesStatesCountMap[*instanceState.State] = instancesStatesCountMap[*instanceState.State] + 1
	}
	if len(instancesStatesCountMap) <= 0 {
		fmt.Println("Load Balance - ", elbName, " does not have any node")
		return
	}
	for state, count := range instancesStatesCountMap {
		message := fmt.Sprintf("%d number of instances are in state %s", count, state)
		fmt.Println(message)
	}
	if critical > 0 && instancesStatesCountMap["InService"] < critical {
		message := fmt.Sprintf("CRITICAL : %d number of instances are in state %s ", instancesStatesCountMap["InService"], "InService")
		fmt.Println(message)
	} else if warning > 0 && instancesStatesCountMap["InService"] < warning {
		message := fmt.Sprintf("WARNING : %d number of instances are in state %s ", instancesStatesCountMap["InService"], "InService")
		fmt.Println(message)
	}
	if criticalPercentage > 0 || warningPercentage > 0 {
		pecentage = float64(instancesStatesCountMap["InService"]) / float64(len(instanceStates))
		pecentage = pecentage * 100
	}
	if criticalPercentage > 0 && pecentage < criticalPercentage {
		message := fmt.Sprintf("CRITICAL : %v percentage are in state %s ", pecentage, "InService")
		fmt.Println(message)
	} else if warningPercentage > 0 && pecentage < warningPercentage {
		message := fmt.Sprintf("WARNING : %v percentage are in state %s ", pecentage, "InService")
		fmt.Println(message)
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
	checkNodes()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-elb-nodes",
		Short: "The Sensu Go Aws Load Balancer handler for node management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsregion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	cmd.Flags().StringVar(&elbName, "load_balancer", "", "The name of the ELB")
	cmd.Flags().IntVar(&warning, "warning", -1, "Minimum number of nodes InService on the ELB to be considered a warning")
	cmd.Flags().IntVar(&critical, "critical", -1, "Minimum number of nodes InService on the ELB to be considered critical")
	cmd.Flags().Float64Var(&warningPercentage, "warning_percentage", -1, "Warn when the percentage of InService nodes is at or below this number")
	cmd.Flags().Float64Var(&criticalPercentage, "critical_percentage", -1, "CRITICAL when the percentage of InService nodes is at or below this number")

	return cmd
}
