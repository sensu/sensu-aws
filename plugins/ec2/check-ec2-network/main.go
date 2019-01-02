package main

/*
# check-ec2-network
#
# DESCRIPTION:
#   Check EC2 Network Metrics by CloudWatch API.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-ec2-network --instance_id=i-0f1626fsbfvbafa2 --direction=NetworkOut
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
	"time"

	"github.com/spf13/cobra"
	"github.com/sreejita-biswas/aws-plugins/awsclient"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	ec2Client         *ec2.EC2
	cloudWatchClient  *cloudwatch.CloudWatch
	criticalThreshold float64
	warningThreshold  float64
	instanceId        string
	endTime           string
	period            int64
	direction         string
	awsRegion         string
)

func checkNetwork() {
	var success bool

	if !(direction == "NetworkIn" || direction == "NetworkOut") {
		fmt.Println("Invalid direction")
		return
	}

	endTimeDate, err := time.Parse(time.RFC3339, endTime)

	if err != nil {
		fmt.Println("Invalid end time entered , ", err)
		return
	}

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, ec2Client = awsclient.GetEC2Client(awsSession)
	if !success {
		return
	}
	success, cloudWatchClient = awsclient.GetCloudWatchClient(awsSession)
	if !success {
		return
	}
	networkValue, err := getEc2NetworkMetric(endTimeDate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if networkValue != nil {
		if *networkValue > criticalThreshold {
			fmt.Println("CRITICAL:", direction, "at", *networkValue, "bytes")
		} else if *networkValue > warningThreshold {
			fmt.Println("WARNINg:", direction, " at ", *networkValue, "bytes")
		} else {
			fmt.Println("OK:", direction, "at", *networkValue, "bytes")
		}
	}

}

func getEc2NetworkMetric(endTimeDate time.Time) (*float64, error) {
	stats := "Average"
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/EC2")
	input.MetricName = aws.String(direction)
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("InstanceId")
	dimensionFilter.Value = aws.String(instanceId)
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter}
	input.EndTime = aws.Time(endTimeDate)
	input.StartTime = aws.Time(endTimeDate.Add(time.Duration(-5) * time.Minute))
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(stats)}
	input.Unit = aws.String("Bytes")
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		return nil, err
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) > 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		var averageValue *float64
		minimumTimeDifference = -1
		for _, datapoint := range metrics.Datapoints {
			timeDifference = time.Since(*datapoint.Timestamp).Seconds()
			if minimumTimeDifference == -1 {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
			} else if timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
			}
		}
		return averageValue, nil
	}
	return nil, nil
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
	checkNetwork()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-ec2-netwrok",
		Short: "The Sensu Go Aws EC2 handler for netwrok management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region")
	cmd.Flags().Float64Var(&criticalThreshold, "critical", 1000000, "Trigger a critical if network traffice is over specified Bytes")
	cmd.Flags().Float64Var(&warningThreshold, "warning", 1500000, "Trigger a warning if network traffice is over specified Bytes")
	cmd.Flags().StringVar(&instanceId, "instance_id", "", "EC2 Instance ID to check.")
	cmd.Flags().StringVar(&endTime, "end_time", time.Now().Format(time.RFC3339), "CloudWatch metric statistics end time, e.g. 2014-11-12T11:45:26.371Z")
	cmd.Flags().Int64Var(&period, "period", 60, "CloudWatch metric statistics period in seconds")
	cmd.Flags().StringVar(&direction, "direction", "NetworkIn", "Select NetworkIn or NetworkOut")

	_ = cmd.MarkFlagRequired("instance_id")
	return cmd
}
