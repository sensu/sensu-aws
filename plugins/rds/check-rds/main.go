package main

/*
#
# check-rds
#
# DESCRIPTION:
#   Check RDS instance statuses by RDS and CloudWatch API.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   Critical if DB instance "sensu-admin-db" is not on ap-northeast-1a
#   ./check-rds --db_instance_id=sensu-admin-db --available_zone_severity=critical --available_zone=ap-northeast-1a
#
#   Warning if CPUUtilization is over 80%, critical if over 90%
#   ./check-rds --db_instance_id=sensu-admin-db --cpu_warning_over=80 --cpu_critical_over=90
#
#   Critical if CPUUtilization is over 90%, maximum of last one hour
#   ./check-rds --db_instance_id=sensu-admin-db --cpu_critical_over=90 --statistics=maximum --period=3600
#
#   Warning if DatabaseConnections are over 100, critical over 120
#   ./check-rds --db_instance_id=sensu-admin-db --connections_critical_over=120 --connections_warning_over=100 --statistics=maximum --period=3600
#
#   Warning if IOPS are over 100, critical over 200
#   ./check-rds --db_instance_id=sensu-admin-db --iops_critical_over=200 --iops_warning_over=100 --period=300
#
#   Warning if memory usage is over 80%, maximum of last 2 hour
#   specifying "minimum" is intended actually since memory usage is calculated from CloudWatch "FreeableMemory" metric.
#   ./check-rds --db_instance_id=sensu-admin-db --memory_warning_over=80 --statistics=minimum --period=7200
#
#   Disk usage, same as memory
#   ./check-rds --db_instance_id=sensu-admin-db --disk_warning_over=80 --period=7200
#
#   You can check multiple metrics simultaneously. Highest severity will be reported
#   ./check-rds --db_instance_id=sensu-admin-db --cpu_warning_over=80 --cpu_critical_over=90 --memory_warning_over=60 --memory_critical_over=80
#
#   You can ignore accept nil values returned for a time periods from Cloudwatch as being an OK.  Amazon falls behind in their
#   metrics from time to time and this prevents false positives
#   ./check-rds --db_instance_id=sensu-admin-db --cpu_critical_over=90 -accept_nil=true
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sreejita-biswas/aws-handler/awsclient"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sreejita-biswas/aws-handler/aws_session"
)

var (
	awsRegion                string
	rdsClient                *rds.RDS
	scheme                   string
	dbInstanceId             string
	fetchAge                 int
	period                   int64
	statistic                string
	cloudWatchClient         *cloudwatch.CloudWatch
	roleArn                  string
	dbClusterId              string
	accpetNil                bool
	availabilityZoneSeverity string
	cpuCritical              float64
	cpuWarning               float64
	memoryCritical           float64
	memoryWarning            float64
	diskCritical             float64
	diskWarning              float64
	conectionCritical        float64
	connectionWarning        float64
	iopsCritical             float64
	iopsWarning              float64
	metricSeverities         map[string]map[string]float64
	availabilityZone         string
	instanceClassMapping     map[string]string
	allocatedStorageMapping  map[string]int64
	dbInstanceZoneMapping    map[string]string
)

func checkRds() {
	var success bool
	metrics := getMetrics()
	metricSeverities = getMetricSeverities()
	values := make(map[string]*float64)
	severities := make(map[string]bool)
	severities["critical"] = false
	severities["warning"] = false
	instanceClassMapping = make(map[string]string)
	allocatedStorageMapping = make(map[string]int64)
	dbInstanceZoneMapping = make(map[string]string)
	if len(dbClusterId) <= 0 && len(dbInstanceId) <= 0 {
		fmt.Println("Please provide db_cluster_id or db_instance_id")
		return
	}
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	if len(roleArn) <= 0 {
		success, rdsClient = awsclient.GetRDSClient(awsSession)
	} else {
		success, rdsClient = awsclient.GetRDSClientWithRoleArn(awsSession, roleArn)
	}
	if !success {
		return
	}
	if len(roleArn) <= 0 {
		success, cloudWatchClient = awsclient.GetCloudWatchClient(awsSession)
	} else {
		success, cloudWatchClient = awsclient.GetCloudWatchClientWithRoleArn(awsSession, roleArn)
	}
	if !success {
		return
	}
	err := getClusterDetails()
	if err != nil {
		return
	}
	err = getDbInstanceDetails()
	if err != nil {
		return
	}
	for instance, zone := range dbInstanceZoneMapping {
		values = make(map[string]*float64)
		//if zone != availabilityZone {
		// if availabilityZoneSeverity == "citical" {
		// 	fmt.Print("CRITICAL :")
		// } else {
		// 	fmt.Print("WARNING :")
		// }
		fmt.Println("Availabilty Zone for DB Instance", instance, "is", zone)
		//}
		for metric, unit := range metrics {
			value, err := getCloudWatchMetrics(metric, instance, unit)
			if err != nil {
				fmt.Println("Error :", err)
				return
			}
			values[metric] = value
		}

		checkCPU(values["CPUUtilization"], instance)
		checkMemory(values["FreeableMemory"], instance, instanceClassMapping[instance])
		checkDiskSpace(values["FreeStorageSpace"], instance, allocatedStorageMapping[instance])
		checkConnections(values["DatabaseConnections"], instance)
		checkIops(values["ReadIOPS"], values["WriteIOPS"], instance)
	}
}

func getCloudWatchMetrics(metricname string, rdsName string, unit string) (*float64, error) {
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/RDS")
	input.MetricName = aws.String(metricname)
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("DBInstanceIdentifier")
	dimensionFilter.Value = aws.String(rdsName)
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter}
	input.EndTime = aws.Time(time.Now())
	input.StartTime = aws.Time((input.EndTime).Add(time.Duration(-period/60) * time.Minute))
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(strings.Title(statistic))}
	input.Unit = aws.String(unit)
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		return nil, err
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) > 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		var value *float64
		minimumTimeDifference = -1
		for _, datapoint := range metrics.Datapoints {
			timeDifference = time.Since(*datapoint.Timestamp).Seconds()
			if minimumTimeDifference == -1 || timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				if strings.Title(statistic) == "Average" {
					value = datapoint.Average
				} else if strings.Title(statistic) == "Sum" {
					value = datapoint.Sum
				} else if strings.Title(statistic) == "Maximum" {
					value = datapoint.Maximum
				} else if strings.Title(statistic) == "Minimum" {
					value = datapoint.Minimum
				}
			}
		}
		return value, nil
	}
	return nil, nil
}

func getMetrics() map[string]string {
	metrics := make(map[string]string)
	metrics["CPUUtilization"] = "Percent"
	metrics["FreeableMemory"] = "Bytes"
	metrics["FreeStorageSpace"] = "Bytes"
	metrics["DatabaseConnections"] = "Count"
	metrics["ReadIOPS"] = "Count/Second"
	metrics["WriteIOPS"] = "Count/Second"
	return metrics
}

func getMetricSeverities() map[string]map[string]float64 {
	metricSeverities := make(map[string]map[string]float64)
	metricSeverities["CPUUtilization"] = make(map[string]float64)
	metricSeverities["CPUUtilization"]["critical"] = cpuCritical
	metricSeverities["CPUUtilization"]["warning"] = cpuWarning
	metricSeverities["FreeableMemory"] = make(map[string]float64)
	metricSeverities["FreeableMemory"]["critical"] = memoryCritical
	metricSeverities["FreeableMemory"]["warning"] = memoryWarning
	metricSeverities["FreeStorageSpace"] = make(map[string]float64)
	metricSeverities["FreeStorageSpace"]["critical"] = diskCritical
	metricSeverities["FreeStorageSpace"]["warning"] = diskWarning
	metricSeverities["DatabaseConnections"] = make(map[string]float64)
	metricSeverities["DatabaseConnections"]["critical"] = conectionCritical
	metricSeverities["DatabaseConnections"]["warning"] = connectionWarning
	metricSeverities["ReadIOPS"] = make(map[string]float64)
	metricSeverities["ReadIOPS"]["critical"] = iopsCritical
	metricSeverities["ReadIOPS"]["warning"] = iopsWarning
	metricSeverities["WriteIOPS"] = make(map[string]float64)
	metricSeverities["WriteIOPS"]["critical"] = iopsCritical
	metricSeverities["WriteIOPS"]["warning"] = iopsWarning
	return metricSeverities
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-rds",
		Short: "The Sensu Go Aws RDS handler for rds management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-west-1", "AWS Region (defaults to us-east-1).")
	cmd.Flags().StringVar(&scheme, "scheme", "", "Metric naming scheme, text to prepend to metric")
	cmd.Flags().StringVar(&dbInstanceId, "db_instance_id", "", "DB instance identifier")
	cmd.Flags().IntVar(&fetchAge, "fetch_age", 0, "How long ago to fetch metrics from in seconds")
	cmd.Flags().Int64Var(&period, "period", 180, "CloudWatch metric statistics period")
	cmd.Flags().StringVar(&statistic, "statistic", "average", "CloudWatch statistics method")
	cmd.Flags().StringVar(&dbClusterId, "db_cluster_id", "", "DB cluster identifier")
	cmd.Flags().BoolVar(&accpetNil, "accept_nil", false, "Continue if CloudWatch provides no metrics for the time period")
	cmd.Flags().StringVar(&availabilityZoneSeverity, "available_zone_severity", "critical", "Trigger a #{severity} if availability zone is different than given argument")
	cmd.Flags().StringVar(&availabilityZone, "available_zone", "us-west-1a", "available zone")
	cmd.Flags().Float64Var(&cpuCritical, "cpu_critical_over", 80, "Trigger a critical if cpu usage is over a percentage")
	cmd.Flags().Float64Var(&cpuWarning, "cpu_warning_over", 40, "Trigger a warning if cpu usage is over a percentage")
	cmd.Flags().Float64Var(&memoryCritical, "memory_critical_over", 80, "Trigger a critical if memory usage is over a Bytes")
	cmd.Flags().Float64Var(&memoryWarning, "memory_warning_over", 40, "Trigger a warning if memory usage is over a Bytes")
	cmd.Flags().Float64Var(&diskCritical, "disk_critical_over", 80, "Trigger a critical if disk usage is over a Bytes")
	cmd.Flags().Float64Var(&diskWarning, "disk_warning_over", 40, "Trigger a warning if disk usage is over a Bytes")
	cmd.Flags().Float64Var(&conectionCritical, "connections_critical_over", 80, "Trigger a critical if connection number is over a number")
	cmd.Flags().Float64Var(&connectionWarning, "connections_warning_over", 40, "Trigger a warning if connection number is over a number")
	cmd.Flags().Float64Var(&iopsCritical, "iops_critical_over", 80, "Trigger a critical if iops number is over a Count/Second")
	cmd.Flags().Float64Var(&iopsWarning, "iops_warning_over", 40, "Trigger a warning if connection number is over a Count/Second")
	cmd.Flags().StringVar(&roleArn, "role_arn", "", "AWS role arn of the role of the third party account to switch to")
	return cmd
}

func checkCPU(value *float64, instance string) {
	if checkNilValue(value, instance, "cpu") {
		return
	}

	if *value >= metricSeverities["CPUUtilization"]["critical"] {
		fmt.Println("CRITICAL : For DB Instance :", instance, "latest cpu usage value :", *value, "expected lower than", metricSeverities["CPUUtilization"]["critical"])
	}

	if *value >= metricSeverities["CPUUtilization"]["warning"] {
		fmt.Println("WARNING : For DB Instance :", instance, "latest cpu usage value :", *value, "expected lower than", metricSeverities["CPUUtilization"]["warning"])
	}
}

func checkMemory(value *float64, instance string, instanceClass string) {
	if checkNilValue(value, instance, "memory") {
		return
	}

	memoryTotalBytes := getMemoryTotalBytes(instanceClass)
	memoryUsageBytes := memoryTotalBytes - *value
	memoryUsagePercentage := (memoryUsageBytes / memoryTotalBytes) * 100

	if memoryUsagePercentage >= metricSeverities["FreeableMemory"]["critical"] {
		fmt.Println("CRITICAL : For DB Instance :", instance, "latest memory usage value :", memoryUsagePercentage, "expected lower than", metricSeverities["FreeableMemory"]["critical"])
	}

	if memoryUsagePercentage >= metricSeverities["FreeableMemory"]["warning"] {
		fmt.Println("WARNING : For DB Instance :", instance, "latest memory usage value :", memoryUsagePercentage, "expected lower than", metricSeverities["FreeableMemory"]["warning"])
	}
}

func checkDiskSpace(value *float64, instance string, allocatedStorage int64) {
	if checkNilValue(value, instance, "disk") {
		return
	}

	diskTotalBytes := float64(allocatedStorage) * math.Pow(1024, 3)
	diskUsageBytes := diskTotalBytes - *value
	diskUsagePercentage := (diskUsageBytes / diskTotalBytes) * 100

	if diskUsagePercentage >= metricSeverities["FreeStorageSpace"]["critical"] {
		fmt.Println("CRITICAL : For DB Instance :", instance, "latest disk usage value :", diskUsagePercentage, "expected lower than", metricSeverities["FreeStorageSpace"]["critical"])
	}

	if diskUsagePercentage >= metricSeverities["FreeStorageSpace"]["warning"] {
		fmt.Println("WARNING : For DB Instance :", instance, "latest disk usage value :", diskUsagePercentage, "expected lower than", metricSeverities["FreeStorageSpace"]["warning"])
	}
}

func checkConnections(value *float64, instance string) {
	if checkNilValue(value, instance, "database connections") {
		return
	}

	if *value >= metricSeverities["DatabaseConnections"]["critical"] {
		fmt.Println("CRITICAL : For DB Instance :", instance, "latest cpu usage value :", *value, "expected lower than", metricSeverities["DatabaseConnections"]["critical"])
	}

	if *value >= metricSeverities["DatabaseConnections"]["warning"] {
		fmt.Println("WARNING : For DB Instance :", instance, "latest cpu usage value :", *value, "expected lower than", metricSeverities["DatabaseConnections"]["warning"])
	}
}

func checkIops(value *float64, value2 *float64, instance string) {
	isReadIopsNull := checkNilValue(value, instance, "iops")

	if isReadIopsNull {
		return
	}

	isWriteIopsNull := checkNilValue(value2, instance, "iops")

	if isWriteIopsNull {
		return
	}

	iopsValue := *value + *value2
	if iopsValue >= metricSeverities["CPUUtilization"]["critical"] {
		fmt.Println("CRITICAL : For DB Instance :", instance, "latest iops usage value :", iopsValue, "expected lower than", metricSeverities["CPUUtilization"]["critical"])
	}

	if iopsValue >= metricSeverities["CPUUtilization"]["warning"] {
		fmt.Println("WARNING : For DB Instance :", instance, "latest iops usage value :", iopsValue, "expected lower than", metricSeverities["CPUUtilization"]["warning"])
	}
}

func checkNilValue(value *float64, instance string, metric string) bool {
	if value == nil && accpetNil {
		fmt.Println("DB INSTACE :", instance, ":", metric, "usage : Cloudwatch returned no results for time period. Accept nil passed so OK")
		return true
	}
	if value == nil && !accpetNil {
		fmt.Println("UNKNOWN : DB INSTACE :", instance, ":", metric, "usage : Requested time period did not return values from Cloudwatch. Try increasing your time period.")
		return true
	}
	return false
}

func getMemoryTotalBytes(instaceClass string) float64 {
	memoryByteMap := make(map[string]float64)
	memoryByteMap["db.cr1.8xlarge"] = 244.0
	memoryByteMap["db.m1.small"] = 1.7
	memoryByteMap["db.m1.medium"] = 3.75
	memoryByteMap["db.m1.large"] = 7.5
	memoryByteMap["db.m1.xlarge"] = 15.0
	memoryByteMap["db.m2.xlarge"] = 17.1
	memoryByteMap["db.m2.2xlarge"] = 34.2
	memoryByteMap["db.m2.4xlarge"] = 68.4
	memoryByteMap["db.m3.medium"] = 3.75
	memoryByteMap["db.m3.large"] = 7.5
	memoryByteMap["db.m3.xlarge"] = 15.0
	memoryByteMap["db.m3.2xlarge"] = 30.0
	memoryByteMap["db.m4.large"] = 8.0
	memoryByteMap["db.m4.xlarge"] = 16.0
	memoryByteMap["db.m4.2xlarge"] = 32.0
	memoryByteMap["db.m4.4xlarge"] = 64.0
	memoryByteMap["db.m4.10xlarge"] = 160.0
	memoryByteMap["db.m4.16xlarge"] = 256.0
	memoryByteMap["db.m5.large"] = 8.0
	memoryByteMap["db.m5.xlarge"] = 16.0
	memoryByteMap["db.m5.2xlarge"] = 32.0
	memoryByteMap["db.m5.4xlarge"] = 64.0
	memoryByteMap["db.m5.12xlarge"] = 192.0
	memoryByteMap["db.m5.24xlarge"] = 384.0
	memoryByteMap["db.r3.large"] = 15.0
	memoryByteMap["db.r3.xlarge"] = 30.5
	memoryByteMap["db.r3.2xlarge"] = 61.0
	memoryByteMap["db.r3.4xlarge"] = 122.0
	memoryByteMap["db.r3.8xlarge"] = 244.0
	memoryByteMap["db.r4.large"] = 15.25
	memoryByteMap["db.r4.xlarge"] = 30.5
	memoryByteMap["db.r4.2xlarge"] = 61.0
	memoryByteMap["db.r4.4xlarge"] = 122.0
	memoryByteMap["db.r4.8xlarge"] = 244.0
	memoryByteMap["db.r4.16xlarge"] = 488.0
	memoryByteMap["db.t1.micro"] = 0.615
	memoryByteMap["db.t2.micro"] = 1.0
	memoryByteMap["db.t2.small"] = 2.0
	memoryByteMap["db.t2.medium"] = 4.0
	memoryByteMap["db.t2.large"] = 8.0
	memoryByteMap["db.t2.xlarge"] = 16.0
	memoryByteMap["db.t2.2xlarge"] = 32.0
	memoryByteMap["db.x1.16xlarge"] = 976.0
	memoryByteMap["db.x1.32xlarge"] = 1952.0
	memoryByteMap["db.x1e.xlarge"] = 122.0
	memoryByteMap["db.x1e.2xlarge"] = 244.0
	memoryByteMap["db.x1e.4xlarge"] = 488.0
	memoryByteMap["db.x1e.8xlarge"] = 976.0
	memoryByteMap["db.x1e.16xlarge"] = 1952.0
	memoryByteMap["db.x1e.32xlarge"] = 3904.0

	return memoryByteMap[instaceClass] * math.Pow(1024, 3)
}

func getClusterDetails() error {
	if len(dbClusterId) > 0 {
		dbclustersInput := &rds.DescribeDBClustersInput{}

		filter := &rds.Filter{}
		filter.Name = aws.String("db-cluster-id")
		filter.Values = []*string{&dbClusterId}
		dbclustersInput.Filters = []*rds.Filter{filter}

		dbClusterOutput, err := rdsClient.DescribeDBClusters(dbclustersInput)

		if err != nil {
			fmt.Println("An error occurred processing AWS RDS API DescribeDBClusters", err)
			return err
		}

		if dbClusterOutput == nil || dbClusterOutput.DBClusters == nil || len(dbClusterOutput.DBClusters) <= 0 {
			fmt.Println("UNKNOWN : DB Cluster not found!")
		}

		if dbClusterOutput != nil && dbClusterOutput.DBClusters != nil && len(dbClusterOutput.DBClusters) > 0 {
			for _, dbclustersMember := range dbClusterOutput.DBClusters[0].DBClusterMembers {
				if *dbclustersMember.IsClusterWriter {
					dbInstanceInput := &rds.DescribeDBInstancesInput{}
					filter := &rds.Filter{}
					filter.Name = aws.String("db-instance-id")
					filter.Values = []*string{aws.String(*dbclustersMember.DBInstanceIdentifier)}
					dbInstanceInput.Filters = []*rds.Filter{filter}

					dbClusterOutput, err := rdsClient.DescribeDBInstances(dbInstanceInput)

					if err != nil {
						fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
						return err
					}

					if dbClusterOutput == nil || dbClusterOutput.DBInstances == nil || len(dbClusterOutput.DBInstances) <= 0 {
						fmt.Println("UNKNOWN :", dbInstanceId, "instance not found")
					} else {
						dbInstanceZoneMapping[*dbclustersMember.DBInstanceIdentifier] = *dbClusterOutput.DBInstances[0].AvailabilityZone
						instanceClassMapping[*dbclustersMember.DBInstanceIdentifier] = *dbClusterOutput.DBInstances[0].DBInstanceClass
						allocatedStorageMapping[*dbclustersMember.DBInstanceIdentifier] = *dbClusterOutput.DBInstances[0].AllocatedStorage
					}
				}
			}
		}
	}
	return nil
}

func getDbInstanceDetails() error {
	if len(dbInstanceId) > 0 {
		dbInstanceInput := &rds.DescribeDBInstancesInput{}
		filter := &rds.Filter{}
		filter.Name = aws.String("db-instance-id")
		filter.Values = []*string{aws.String(dbInstanceId)}
		dbInstanceInput.Filters = []*rds.Filter{filter}

		dbClusterOutput, err := rdsClient.DescribeDBInstances(dbInstanceInput)

		if err != nil {
			fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
			return err
		}

		if dbClusterOutput == nil || dbClusterOutput.DBInstances == nil || len(dbClusterOutput.DBInstances) <= 0 {
			fmt.Println("UNKNOWN :", dbInstanceId, "instance not found")
		} else {
			dbInstanceZoneMapping[dbInstanceId] = *dbClusterOutput.DBInstances[0].AvailabilityZone
			instanceClassMapping[dbInstanceId] = *dbClusterOutput.DBInstances[0].DBInstanceClass
			allocatedStorageMapping[dbInstanceId] = *dbClusterOutput.DBInstances[0].AllocatedStorage
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
	checkRds()
	return nil
}
