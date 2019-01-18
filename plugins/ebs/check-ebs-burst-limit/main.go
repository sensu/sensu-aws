package main

/*
#
# check-ebs-burst-limit
#
# DESCRIPTION:
#   Check EC2 Volumes for volumes with low burst balance
#   Optionally check only volumes attached to the current instance
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-ebs-burst-limit
#
# LICENSE:
#   TODO
#
*/

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-handler/aws_session"
	"github.com/sreejita-biswas/aws-handler/awsclient"
)

var (
	ec2Client         *ec2.EC2
	scheme            string
	awsRegion         string
	cloudWatchClient  *cloudwatch.CloudWatch
	criticalThreshold float64
	warningThreshold  float64
	checkSelf         bool
)

func checkLimit() {
	var success bool
	volumeInput := &ec2.DescribeVolumesInput{}

	// Set the describe-volumes filter depending on whether -s was specified
	if checkSelf {
		//TODO
	} else {
		//The --check_self option was not specified, look at all volumes which are attached
		filter := &ec2.Filter{}
		filter.Name = aws.String("attachment.status")
		filter.Values = []*string{aws.String("attached")}
		volumeInput.Filters = []*ec2.Filter{filter}
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

	volumes, err := ec2Client.DescribeVolumes(volumeInput)
	if err != nil {
		fmt.Println(err)
	}

	isCritical := false
	shouldWarn := false
	errors := []string{}

	if volumes != nil {
		for _, volume := range volumes.Volumes {
			crit, warn, errorString := getMetric(*volume.VolumeId)
			if errorString != nil && len(errorString) > 0 {
				errors = append(errors, errorString[0])
			}
			isCritical = isCritical || crit
			shouldWarn = shouldWarn || warn
		}
	}

	if isCritical {
		fmt.Println("CRITICAL : Volume(s) have exceeded critical threshold:", errors)
	} else if shouldWarn {
		fmt.Println("WARNING : Volume(s) have exceeded warning threshold:", errors)
	}
}

func getMetric(volumeId string) (bool, bool, []string) {
	errors := []string{}
	isCRitical := false
	shouldWarn := false
	stats := "Average"
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/EBS")
	input.MetricName = aws.String("BurstBalance")
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("VolumeId")
	dimensionFilter.Value = aws.String(volumeId)
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter}
	input.Period = aws.Int64(120)
	input.EndTime = aws.Time(time.Now())
	input.StartTime = aws.Time(time.Now().Add(time.Duration(-24*60) * time.Minute))
	input.Statistics = []*string{aws.String(stats)}
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		fmt.Println(err)
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) >= 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		var averageValue *float64
		minimumTimeDifference = -1
		for _, datapoint := range metrics.Datapoints {
			timeDifference = time.Since(*datapoint.Timestamp).Seconds()
			if minimumTimeDifference == -1 || timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
			}
		}
		if *averageValue < criticalThreshold {
			errors = append(errors, fmt.Sprintf("%v:%v", volumeId, averageValue))
			isCRitical = true
		} else if *averageValue < warningThreshold {
			errors = append(errors, fmt.Sprintf("%v:%v", volumeId, averageValue))
			shouldWarn = true
		}
	}
	return isCRitical, shouldWarn, errors
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
	checkLimit()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-ebs-burst-limit",
		Short: "The Sensu Go Aws EBS handler for burst limit management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-2", "AWS Region (defaults to us-east-1).")
	cmd.Flags().Float64Var(&criticalThreshold, "critical", 50, "Trigger a critical when ebs burst limit is under VALUE")
	cmd.Flags().Float64Var(&warningThreshold, "warning", 10, "Trigger a warning when ebs burst limit is under VALUE")
	cmd.Flags().BoolVar(&checkSelf, "check_self", false, "Only check the instance on which this plugin is being run - this overrides the -r option and uses the region of the current instance")

	return cmd
}
