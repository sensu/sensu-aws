package main

/*
#
# metrics-ec2-filter
#
# DESCRIPTION:
#   This plugin retrieves EC2 instances matching a given filter and
#   returns the number matched. Warning and Critical thresholds may be set as needed.
#   Thresholds may be compared to the count using [equal, not, greater, less] operators.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./metric-ec2-filter --filters="{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}"
# NOTES:
#
# LICENSE:
#   TODO
#
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sensu/sensu-aws/aws_session"
	"github.com/sensu/sensu-aws/awsclient"
	"github.com/sensu/sensu-aws/models"
	"github.com/sensu/sensu-aws/utils"
	"github.com/spf13/cobra"
)

var (
	ec2Client  *ec2.EC2
	filters    string
	metricType string
	scheme     string
	filterName string
	awsRegion  string
)

func metrics() {
	var success bool
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

	success, ec2Client = awsclient.GetEC2Client(awsSession)
	if !success {
		return
	}
	var ec2Fileters models.Filters
	err := json.Unmarshal([]byte(filters), &ec2Fileters)
	if err != nil {
		fmt.Println("Failed to unmarshal filter data , ", err)
		return
	}

	reservations, err := utils.GetReservations(ec2Client, ec2Fileters.Filters)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Print("EC2 Instances of ", scheme)
	if len(strings.TrimSpace(filterName)) > 1 {
		fmt.Println("with filter :", filterName)
	}
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			fmt.Println(*instance.InstanceId)
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
		Use:   "metrics-ec2-filter",
		Short: "The Sensu Go Aws EC2 handler for instance filter management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "Aws Region")
	cmd.Flags().StringVar(&metricType, "metric_type", "instance", "Count by type: status, instance")
	cmd.Flags().StringVar(&scheme, "scheme", "sensu.aws.ec2", "Metric naming scheme, text to prepend to metric")
	cmd.Flags().StringVar(&filters, "filters", "{}", "JSON String representation of Filters, e.g. {\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}")
	cmd.Flags().StringVar(&filterName, "filter_name", "", "Filter naming scheme, text to prepend to metric")

	return cmd
}
