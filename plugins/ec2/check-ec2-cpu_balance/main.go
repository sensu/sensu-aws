package main

/* check-ec2-cpu_balance
#
# DESCRIPTION:
#   This plugin retrieves the value of the cpu balance for all servers
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-ec2-cpu_balance --critical=3
#   ./check-ec2-cpu_balance --critical=1 --warning=5
#   ./check-ec2-cpu_balance --critical=1 --warning=5 --tag=TESTING
#
# NOTES:
#
# LICENSE:
#   TODO
*/

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sreejita-biswas/aws-plugins/awsclient"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
	"github.com/sreejita-biswas/aws-plugins/utils"
)

var (
	ec2Client         *ec2.EC2
	cloudWatchClient  *cloudwatch.CloudWatch
	criticalThreshold float64
	warningThreshold  float64
	tagValue          string
	awsRegion         string
)

func ckeckCpu() {
	var success bool
	var reservations []*ec2.Reservation
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, ec2Client = awsclient.GetEC2Client(awsSession)
	if !success {
		return
	}
	filter := ec2.Filter{Name: aws.String("instance-state-name"), Values: []*string{
		aws.String("running")}}

	reservations, err := utils.GetReservations(ec2Client, []*ec2.Filter{&filter})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	success, cloudWatchClient = awsclient.GetCloudWatchClient(awsSession)
	if !success {
		return
	}

	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			instanceType := instance.InstanceType
			if strings.HasPrefix(*instanceType, "t2.") {
				cpuBalance, err := getEc2CpuBalance(*instance)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				tagValue := getMatchingInstanceTag(*instance)
				if tagValue != nil {
					if *cpuBalance < criticalThreshold {
						fmt.Println(*instance.InstanceId, *tagValue, "is below critical threshold", "[cpuBalance <", criticalThreshold, "]")
					} else if *cpuBalance < warningThreshold {
						fmt.Println(*instance.InstanceId, *tagValue, "is below warning threshold", "[cpuBalance <", warningThreshold, "]")
					}
				}
			}
		}
	}
}

func getEc2CpuBalance(instance ec2.Instance) (*float64, error) {
	stats := "Average"
	var period int64
	period = 60
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/EC2")
	input.MetricName = aws.String("CPUCreditBalance")
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("InstanceId")
	dimensionFilter.Value = instance.InstanceId
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter}
	input.EndTime = aws.Time(time.Now())
	input.StartTime = aws.Time(time.Now().Add(time.Duration(-10*(period/60)) * time.Minute))
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(stats)}
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		return nil, err
	}
	if metrics != nil {
		for _, datapoint := range metrics.Datapoints {
			return datapoint.Average, nil
		}

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

func getMatchingInstanceTag(instance ec2.Instance) *string {
	for _, tag := range instance.Tags {
		if *tag.Key == tagValue {
			return tag.Value
		}
	}
	return nil
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
	ckeckCpu()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-ec2-cpu_balance",
		Short: "The Sensu Go Aws EC2 handler for cpu management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS region")
	cmd.Flags().Float64Var(&criticalThreshold, "critical", 1.2, "Trigger a critical when value is below the criticalThreshold.")
	cmd.Flags().Float64Var(&warningThreshold, "warning", 2.3, "Trigger a warning when value is below warningThreshold")
	cmd.Flags().StringVar(&tagValue, "tag", "NAME", "Add instance TAG value to warn/critical message.")

	return cmd
}
