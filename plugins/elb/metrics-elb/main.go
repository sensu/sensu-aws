package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/sreejita-biswas/aws-plugins/awsclient"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

/*
#
# elb-metrics
#
# DESCRIPTION:
#   Gets latency metrics from CloudWatch and puts them in Graphite for longer term storage
#
# OUTPUT:
#   metric-data
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   #YELLOW
#
# NOTES:
#   Returns latency statistics by default.  You can specify any valid ELB metric type, see
#   http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CW_Support_For_AWS.html#elb-metricscollected
#
#   By default fetches statistics from one minute ago.  You may need to fetch further back than this;
#   high traffic ELBs can sometimes experience statistic delays of up to 10 minutes.  If you experience this,
#   raising a ticket with AWS support should get the problem resolved.
#   As a workaround you can use eg -f 300 to fetch data from 5 minutes ago.
#
# LICENSE:
#   TODO
*/

var (
	awsRegion    string
	elbName      string
	period       int64
	criticalOver float64
	warningOver  float64
	fetchAge     int64
	//scheme           string
	elbClient        *elb.ELB
	ec2Client        *ec2.EC2
	cloudWatchClient *cloudwatch.CloudWatch
)

func metrics() {
	var awsSession *session.Session
	var success bool
	var elb *string
	awsSession = aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, elbClient = awsclient.GetElbClient(awsSession)
	if !success {
		return
	}
	if len(elbName) > 0 {
		elb = &elbName
	} else {
		elb = nil
	}
	success, elbs := getLoadBalancers(elb)
	if !success {
		return
	}
	success, cloudWatchClient = awsclient.GetCloudWatchClient(awsSession)
	if !success {
		return
	}
	metrics := getMetrics()
	for _, loadBalancer := range elbs {
		for _, metric := range metrics {
			printStatistic(loadBalancer, metric, getMetricStatisticMapping(metric))
		}
	}
}

func getLoadBalancers(elbName *string) (bool, []string) {
	selectedElbs := []string{}
	input := &elb.DescribeLoadBalancersInput{}
	if elbName != nil {
		input.LoadBalancerNames = []*string{elbName}
	}
	output, err := elbClient.DescribeLoadBalancers(input)
	if err != nil {
		fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
		return false, nil
	}

	if !(output != nil && output.LoadBalancerDescriptions != nil && len(output.LoadBalancerDescriptions) > 0) {
		fmt.Println("No Load Balancer found in region -", awsRegion)
		return false, nil
	}

	for _, loadBalancer := range output.LoadBalancerDescriptions {
		selectedElbs = append(selectedElbs, *loadBalancer.LoadBalancerName)
	}
	return true, selectedElbs
}

func printStatistic(elb string, metricName string, statistic string) {
	metricInput := &cloudwatch.GetMetricStatisticsInput{}
	metricInput.Namespace = aws.String("AWS/ELB")
	metricInput.MetricName = aws.String(metricName)
	dimension := &cloudwatch.Dimension{}
	dimension.Name = aws.String("LoadBalancerName")
	dimension.Value = &elb
	metricInput.Dimensions = []*cloudwatch.Dimension{dimension}
	metricInput.EndTime = aws.Time(time.Now().Add(time.Duration(-fetchAge/60) * time.Minute))
	metricInput.StartTime = aws.Time((*metricInput.EndTime).Add(time.Duration(-period/60) * time.Minute))
	metricInput.Statistics = []*string{&statistic}
	metricInput.Period = aws.Int64(period)
	metrics, err := cloudWatchClient.GetMetricStatistics(metricInput)
	if err != nil {
		fmt.Println("Error while getting", statistic, "value for metric", metricName)
		return
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) > 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		var value *float64
		var timestamp time.Time
		minimumTimeDifference = -1
		for _, datapoint := range metrics.Datapoints {
			timeDifference = time.Since(*datapoint.Timestamp).Seconds()
			if minimumTimeDifference == -1 {
				minimumTimeDifference = timeDifference
				if statistic == "Average" {
					value = datapoint.Average
				} else if statistic == "Sum" {
					value = datapoint.Sum
				} else if statistic == "Maximum" {
					value = datapoint.Maximum
				}
				timestamp = *datapoint.Timestamp
			} else if timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				if statistic == "Average" {
					value = datapoint.Average
				} else if statistic == "Sum" {
					value = datapoint.Sum
				} else if statistic == "Maximum" {
					value = datapoint.Maximum
				}
				timestamp = *datapoint.Timestamp
			}
		}
		fmt.Println("Load Balancer :", elb, ", Statistic :", statistic, ", Metric :", metricName, ", Latest Value :", value, ", Timestamp : ", timestamp.Format(time.RFC3339))
	}
}

func getMetricStatisticMapping(metricName string) string {
	metricStatisticMapping := make(map[string]string)
	metricStatisticMapping["Latency"] = "Average"
	metricStatisticMapping["RequestCount"] = "Sum"
	metricStatisticMapping["UnHealthyHostCount"] = "Average"
	metricStatisticMapping["HealthyHostCount"] = "Average"
	metricStatisticMapping["HTTPCode_Backend_2XX"] = "Sum"
	metricStatisticMapping["HTTPCode_Backend_3XX"] = "Sum"
	metricStatisticMapping["HTTPCode_Backend_4XX"] = "Sum"
	metricStatisticMapping["HTTPCode_Backend_5XX"] = "Sum"
	metricStatisticMapping["HTTPCode_ELB_4XX"] = "Sum"
	metricStatisticMapping["HTTPCode_ELB_5XX"] = "Sum"
	metricStatisticMapping["BackendConnectionErrors"] = "Sum"
	metricStatisticMapping["SurgeQueueLength"] = "Maximum"
	metricStatisticMapping["SpilloverCount"] = "Sum"
	return metricStatisticMapping[metricName]
}

func getMetrics() []string {
	metrics := []string{"Latency", "RequestCount",
		"UnHealthyHostCount", "HealthyHostCount", "HTTPCode_Backend_2XX", "HTTPCode_Backend_3XX",
		"HTTPCode_Backend_4XX", "HTTPCode_Backend_5XX", "HTTPCode_ELB_4XX", "HTTPCode_ELB_5XX",
		"BackendConnectionErrors", "SurgeQueueLength", "SpilloverCount"}
	return metrics
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
		Use:   "check-elb-metrics",
		Short: "The Sensu Go Aws Load Balancer handler for metrics management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "eu-east-1", "AWS Region (defaults to us-east-1).")
	cmd.Flags().StringVar(&elbName, "elb_name", "", "Name of the Elastic Load Balancer")
	//flag.StringVar(&scheme, "scheme", "elb", "Metric naming scheme, text to prepend to metric")
	cmd.Flags().Int64Var(&period, "period", 60, "CloudWatch metric statistics period")
	cmd.Flags().Int64Var(&fetchAge, "fetch_age", 60, "How long ago to fetch metrics for in seconds")
	cmd.Flags().Float64Var(&criticalOver, "critical_over", 60, "Trigger a critical severity if latancy is over specified seconds")
	cmd.Flags().Float64Var(&warningOver, "warning_over", 60, "Trigger a warning severity if latancy is over specified seconds")

	return cmd
}
