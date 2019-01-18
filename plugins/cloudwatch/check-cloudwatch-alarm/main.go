package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/sensu/sensu-aws/aws_session"
	"github.com/sensu/sensu-aws/awsclient"
)

/*
#
# check-cloudwatch-alarms
#
# DESCRIPTION:
#   This plugin raise a critical if one of cloud watch alarms are in given state.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./check-cloudwatch-alarm --aws_region=eu-west-1
#   ./check-cloudwatch-alarm --state=ALEARM
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

var (
	excludeAlarms    string
	state            string
	cloudWatchClient *cloudwatch.CloudWatch
	awsRegion        string
)

func checkAlarms() {
	var success bool

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, cloudWatchClient = awsclient.GetCloudWatchClient(awsSession)
	if !success {
		return
	}

	describeInput := &cloudwatch.DescribeAlarmsInput{}
	describeInput.StateValue = aws.String(state)

	describeOutput, err := cloudWatchClient.DescribeAlarms(describeInput)

	if err != nil {
		fmt.Println("Failed to get cloudwatch alarm details , Error : ", err)
	}

	if describeOutput == nil || describeOutput.MetricAlarms == nil || len(describeOutput.MetricAlarms) == 0 {
		fmt.Println("OK : No alarm in", state, "state")
		return
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
	checkAlarms()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-cloudwatch-alarms",
		Short: "The Sensu Go Aws Cloudwatch handler for alarms management",
		RunE:  run,
	}
	cmd.Flags().StringVar(&state, "state", "ALARM", "State of the alarm")
	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	return cmd
}
