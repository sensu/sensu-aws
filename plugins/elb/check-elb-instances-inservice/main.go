package main

/*
#
# check-elb-instance-inservice
#
# DESCRIPTION:
#   Check Elastic Loudbalancer Instances are inService.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   all LoadBalancers
#   ./check-elb-instance-inservice --aws_region=${your_region}
#   one loadBalancer
#   ./check-elb-instance-inservice --aws_region=${your_region} --elb_name=${LoadBalancerName}
#
# NOTES:
#   Based heavily on Peter Hoppe check-autoscaling-instances-inservices
#
# LICENSE:
#  TODO
#
*/

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"

	"github.com/sreejita-biswas/aws-plugins/awsclient"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	awsRegion string
	elbName   string
	elbClient *elb.ELB
	ec2Client *ec2.EC2
)

func checkStatus() {
	var success bool
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, elbClient = awsclient.GetElbClient(awsSession)
	if !success {
		return
	}
	success, ec2Client = awsclient.GetEC2Client(awsSession)
	if !success {
		return
	}
	loadBalancers, err := getLoadBalancers()
	if err != nil || loadBalancers == nil {
		return
	}
	checkInstanceHealth(loadBalancers)
}

func getLoadBalancers() ([]*elb.LoadBalancerDescription, error) {
	input := &elb.DescribeLoadBalancersInput{}
	if len(elbName) > 0 {
		input.LoadBalancerNames = []*string{&elbName}
	}
	output, err := elbClient.DescribeLoadBalancers(input)
	if err != nil {
		fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
		return nil, err
	}

	if !(output != nil && output.LoadBalancerDescriptions != nil && len(output.LoadBalancerDescriptions) > 0) {
		fmt.Println("No Load Balancer found in region -", awsRegion)
		return nil, nil
	}
	return output.LoadBalancerDescriptions, nil
}

func checkInstanceHealth(loadBalancers []*elb.LoadBalancerDescription) {
	for _, loadBalancer := range loadBalancers {
		unhealthyInstances := make(map[string]string)
		instanceStates, err := getHealthStatus(*loadBalancer.LoadBalancerName)
		if err != nil || instanceStates == nil {
			continue
		}
		for _, instanceState := range instanceStates {
			if *instanceState.State != "InService" {
				unhealthyInstances[*instanceState.InstanceId] = *instanceState.State
			}
		}
		if len(unhealthyInstances) == 0 {
			fmt.Println("OK : All instances of Load Balancer - ", *loadBalancer.LoadBalancerName, "are in healthy state")
			continue
		}
		if len(unhealthyInstances) == len(instanceStates) {
			fmt.Println("CRITICAL : All instances of Load Balancer - ", *loadBalancer.LoadBalancerName, "are in unhealthy state")
			continue
		}
		fmt.Println("WARNING : Unhealthy Instances for Load Balanacer - ", *loadBalancer.LoadBalancerName, "are")
		for id, state := range unhealthyInstances {
			fmt.Println("Instace - ", id, ":: State - ", state)
		}

	}
}

func getHealthStatus(elbName string) ([]*elb.InstanceState, error) {
	healtStatusInput := &elb.DescribeInstanceHealthInput{}
	healtStatusInput.LoadBalancerName = aws.String(elbName)
	healtStatusOutput, err := elbClient.DescribeInstanceHealth(healtStatusInput)
	if err != nil {
		fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
		return nil, err
	}
	if !(healtStatusOutput != nil && healtStatusOutput.InstanceStates != nil && len(healtStatusOutput.InstanceStates) > 0) {
		return nil, nil
	}
	return healtStatusOutput.InstanceStates, nil
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
	checkStatus()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-elb-instances-inservice",
		Short: "The Sensu Go Aws Load Balancer handler for instance state management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "eu-west-1", "AWS Region (such as eu-west-1). If you do not specify a region, it will be detected by the server the script is run on")
	cmd.Flags().StringVar(&elbName, "elb_name", "", "The Elastic Load Balancer name of which you want to check the health")

	return cmd
}
