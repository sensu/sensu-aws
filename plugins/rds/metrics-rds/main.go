package main

/*
#
# rds-metrics
#
# DESCRIPTION:
#   Gets RDS metrics from CloudWatch and puts them in Graphite for longer term storage
#
# OUTPUT:
#   metric-data
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./rds-metrics --aws_region=eu-west-1
#   ./rds-metrics --aws_region=eu-west-1 --db_instance_id=sr2x8pbti0eon1
#
# NOTES:
#   Returns all RDS statistics for all RDS instances in this account unless you specify --db_instance_id
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
	"github.com/sreejita-biswas/aws-handler/awsclient"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sreejita-biswas/aws-handler/aws_session"
)

var (
	awsRegion    string
	rdsClient    *rds.RDS
	scheme       string
	dbInstanceId string
	fetchAge     int
	period       int64
	//statistics       string
	cloudWatchClient *cloudwatch.CloudWatch
)

func metrics() {
	var success bool
	clusters := []*string{}
	statisticsTypeMap := getStatisticTypes()
	if len(dbInstanceId) > 0 {
		clusters = []*string{&dbInstanceId}
	}
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, rdsClient = awsclient.GetRDSClient(awsSession)
	if !success {
		return
	}
	dbInstances, err := getDBInstances(clusters)
	if err != nil || dbInstances == nil {
		return
	}
	for _, dbInstance := range dbInstances {
		fullScheme := *dbInstance.DBInstanceIdentifier
		if len(scheme) > 0 {
			fullScheme = fmt.Sprintf("%s.%s", scheme, fullScheme)
		}

		success, cloudWatchClient = awsclient.GetCloudWatchClient(awsSession)
		if !success {
			return
		}
		for metricName, statistic := range statisticsTypeMap {
			value, timestamp, err := getCloudWatchMetrics(metricName, statistic, fullScheme)
			if err != nil {
				fmt.Println("Error : ", err)
				return
			}
			if value == nil || timestamp == nil {
				continue
			}
			fmt.Println(fullScheme, ".", metricName, "  -  value :", value, ", timestamp:", timestamp)
		}
	}
}

func getCloudWatchMetrics(metricname string, statistic string, rdsName string) (*float64, *time.Time, error) {
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/RDS")
	input.MetricName = aws.String(metricname)
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("DBInstanceIdentifier")
	dimensionFilter.Value = aws.String(rdsName)
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter}
	input.EndTime = aws.Time(time.Now().Add(time.Duration(-fetchAge/60) * time.Minute))
	input.StartTime = aws.Time((*input.EndTime).Add(time.Duration(-period/60) * time.Minute))
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(statistic)}
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		return nil, nil, err
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) > 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		var averageValue *float64
		var timestamp *time.Time
		minimumTimeDifference = -1
		for _, datapoint := range metrics.Datapoints {
			timeDifference = time.Since(*datapoint.Timestamp).Seconds()
			if minimumTimeDifference == -1 {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
				timestamp = datapoint.Timestamp
			} else if timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
				timestamp = datapoint.Timestamp
			}
		}
		return averageValue, timestamp, nil
	}
	return nil, nil, nil
}

func getStatisticTypes() map[string]string {
	statisticType := "Average"
	statisticsTypeMap := make(map[string]string)
	statisticsTypeMap["CPUUtilization"] = statisticType
	statisticsTypeMap["DatabaseConnections"] = statisticType
	statisticsTypeMap["FreeStorageSpace"] = statisticType
	statisticsTypeMap["ReadIOPS"] = statisticType
	statisticsTypeMap["ReadLatency"] = statisticType
	statisticsTypeMap["ReadThroughput"] = statisticType
	statisticsTypeMap["WriteIOPS"] = statisticType
	statisticsTypeMap["WriteLatency"] = statisticType
	statisticsTypeMap["WriteThroughput"] = statisticType
	statisticsTypeMap["ReplicaLag"] = statisticType
	statisticsTypeMap["SwapUsage"] = statisticType
	statisticsTypeMap["BinLogDiskUsage"] = statisticType
	statisticsTypeMap["DiskQueueDepth"] = statisticType
	return statisticsTypeMap
}

func getDBInstances(clusters []*string) ([]*rds.DBInstance, error) {
	dbInstanceInput := &rds.DescribeDBInstancesInput{}
	if len(clusters) > 0 {
		filter := &rds.Filter{}
		filter.Name = aws.String("db-instance-id")
		filter.Values = clusters
		dbInstanceInput.Filters = []*rds.Filter{filter}
	}
	dbClusterOutput, err := rdsClient.DescribeDBInstances(dbInstanceInput)

	if err != nil {
		fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
		return nil, err
	}

	if !(dbClusterOutput != nil && dbClusterOutput.DBInstances != nil && len(dbClusterOutput.DBInstances) > 0) {
		fmt.Println("UNKNOWN : DB Instance not found!")
		return nil, nil
	}
	return dbClusterOutput.DBInstances, nil
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
		Use:   "metrics-rds",
		Short: "The Sensu Go Aws RDS handler for metric management",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&awsRegion,
		"aws_region",
		"r",
		"us-east-1",
		"AWS Region")

	cmd.Flags().StringVarP(&scheme,
		"scheme",
		"s",
		"",
		"Metric naming scheme, text to prepend to metric")

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	cmd.Flags().StringVar(&scheme, "scheme", "", "Metric naming scheme, text to prepend to metric")
	cmd.Flags().StringVar(&dbInstanceId, "db_instance_id", "", "DB instance identifier")
	cmd.Flags().IntVar(&fetchAge, "fetch_age", 0, "How long ago to fetch metrics from in seconds")
	cmd.Flags().Int64Var(&period, "period", 60, "CloudWatch metric statistics period")
	return cmd
}
