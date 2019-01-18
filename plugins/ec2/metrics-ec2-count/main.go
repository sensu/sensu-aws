package main

/*
# metrics-ec2-count
#
# DESCRIPTION:
#   This plugin retrieves number of EC2 instances.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   # get metrics on the status of all instances in the region
#   ./metrics-ec2-count.go --metric_type=status
#
#   # get metrics on all instance types in the region
#   ./metrics-ec2-count.go --metric_type=instance
#
# NOTES:
#
# LICENSE:
#  TODO
*/

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/awsclient"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sensu/sensu-aws/aws_session"
	"github.com/sensu/sensu-aws/utils"
)

var (
	ec2Client        *ec2.EC2
	cloudWatchClient *cloudwatch.CloudWatch
	metricType       string
	scheme           string
	awsRegion        string
)

func metrics() {
	var success bool
	metricCount := make(map[string]int)

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, ec2Client = awsclient.GetEC2Client(awsSession)
	if !success {
		return
	}
	success, cloudWatchClient = awsclient.GetCloudWatchClient(awsSession)
	if !success {
		return
	}
	reservations, err := utils.GetReservations(ec2Client, nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	if reservations == nil || len(reservations) <= 0 {
		return
	}

	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			if metricType == "status" {
				metricCount[*instance.State.Name] = metricCount[*instance.State.Name] + 1
			}
			if metricType == "instance" {
				metricCount[*instance.InstanceType] = metricCount[*instance.InstanceType] + 1
			}
		}
	}

	if len(metricCount) > 0 {
		fmt.Println("Number of", scheme, "instances by", metricType)
		for metric, count := range metricCount {
			fmt.Println(metric, "-", count)
		}
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
	metrics()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics-ec2-count",
		Short: "The Sensu Go Aws EC2 handler for number of instance management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region")
	cmd.Flags().StringVar(&metricType, "metric_type", "instance", "Count by type: status, instance")
	cmd.Flags().StringVar(&scheme, "scheme", "sensu.aws.ec2", "Metric naming scheme, text to prepend to metric")

	return cmd
}
