package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sensu/sensu-aws/awsclient"
	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sensu/sensu-aws/aws_session"
)

/*
#
# check-elb-sum-requests
#
# DESCRIPTION:
#   Check ELB Sum Requests by CloudWatch API.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   Warning if any load balancer's sum request count is over 1000, critical if over 2000.
#   ./check-elb-sum-requests --warning_over=1000 --critical_over=2000
#
#   Critical if "app" load balancer's sum request count is over 10000, within last one hour
#   check-elb-sum-requests --elb_names=app --critical_over=10000 --period=3600
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

var (
	awsRegion        string
	elbNames         string
	period           int64
	criticalOver     float64
	warningOver      float64
	elbClient        *elb.ELB
	ec2Client        *ec2.EC2
	cloudWatchClient *cloudwatch.CloudWatch
)

func checkSum() {
	var awsSession *session.Session
	var success bool
	noOfHealthyElbs := 0
	//aws session
	awsSession = aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, elbClient = awsclient.GetElbClient(awsSession)
	if !success {
		return
	}
	success, elbs := getLoadBalancers()
	if !success {
		return
	}
	_, cloudWatchClient = awsclient.GetCloudWatchClient(awsSession)
	for _, elb := range elbs {
		value, startTime, endTime, err := getMetrics(elb)
		if err != nil {
			fmt.Println("Error while getting metrics for Load Balancer - ", elb, ", Error is ", err)
			return
		}
		if value != nil {
			checkSumRequest(*value, elb, *startTime, *endTime)
			continue
		}
		noOfHealthyElbs++
	}
	if noOfHealthyElbs > 0 && noOfHealthyElbs == len(elbs) {
		fmt.Println("OK : ALL load balancers are running with expected sum request value")
	}
}

func getLoadBalancers() (bool, []string) {
	selectedElbs := []string{}
	input := &elb.DescribeLoadBalancersInput{}
	elbs := strings.Split(elbNames, ",")

	elbMap := make(map[string]*string)

	for _, elbName := range elbs {
		elbMap[elbName] = &elbName
	}

	noOfElbs := len(elbMap)

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
		if noOfElbs > 0 && elbMap[*loadBalancer.LoadBalancerName] != nil {
			selectedElbs = append(selectedElbs, *loadBalancer.LoadBalancerName)
		}
	}
	return true, selectedElbs
}

func getMetrics(elb string) (*float64, *string, *string, error) {
	metricInput := &cloudwatch.GetMetricStatisticsInput{}
	metricInput.Namespace = aws.String("AWS/ELB")
	metricInput.MetricName = aws.String("RequestCount")
	dimension := &cloudwatch.Dimension{}
	dimension.Name = aws.String("LoadBalancerName")
	dimension.Value = &elb
	metricInput.Dimensions = []*cloudwatch.Dimension{dimension}
	metricInput.EndTime = aws.Time(time.Now())
	metricInput.StartTime = aws.Time((*metricInput.EndTime).Add(time.Duration(-period/60) * time.Minute))
	metricInput.Statistics = []*string{aws.String("Sum")}
	metricInput.Period = aws.Int64(period)
	metrics, err := cloudWatchClient.GetMetricStatistics(metricInput)
	if err != nil {
		return nil, nil, nil, err
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) > 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		var sumValue *float64
		minimumTimeDifference = -1
		for _, datapoint := range metrics.Datapoints {
			timeDifference = time.Since(*datapoint.Timestamp).Seconds()
			if minimumTimeDifference == -1 {
				minimumTimeDifference = timeDifference
				sumValue = datapoint.Sum
			} else if timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				sumValue = datapoint.Sum
			}
		}
		startTime := metricInput.StartTime.Format(time.RFC3339)
		endTime := metricInput.EndTime.Format(time.RFC3339)
		return sumValue, &startTime, &endTime, nil
	}
	return nil, nil, nil, nil
}

//check latency threshold
func checkSumRequest(value float64, elb string, startTime string, endTime string) {
	if value < criticalOver {
		fmt.Println("CRTICAL : Sum Request for Load Balancer - ", elb, " between ", startTime, " and ", endTime, " is ", value, "(expected lower than ", criticalOver, ")")
		return
	}
	if value < warningOver {
		fmt.Println("WARNING : Sum Request for Load Balancer - ", elb, " between ", startTime, " and ", endTime, " is ", value, "(expected lower than ", criticalOver, ")")
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
	checkSum()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-elb-sum-requests",
		Short: "The Sensu Go Aws Load Balancer handler for sum request management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	cmd.Flags().StringVar(&elbNames, "elb_names", "", "Load balancer names to check. Separated by ,. If not specified, check all load balancers")
	cmd.Flags().Int64Var(&period, "period", 60, "CloudWatch metric statistics period")
	cmd.Flags().Float64Var(&criticalOver, "critical_over", 60, "Trigger a critical severity if latancy is over specified seconds")
	cmd.Flags().Float64Var(&warningOver, "warning_over", 60, "Trigger a warning severity if latancy is over specified seconds")

	return cmd
}
