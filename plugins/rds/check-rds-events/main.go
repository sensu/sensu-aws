package main

/*
#
# check-rds-events
#
#
# DESCRIPTION:
#   This plugin checks rds clusters for critical events.
#   Due to the number of events types on RDS clusters, the check
#   should filter out non-disruptive events that are part of
#   basic operations.
#
#   More info on RDS events:
#   http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Events.html
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#  Check's a specific RDS instance in a specific region for critical events
#  ./check-rds-events --aws_region=${your_region}  --db_instance_id=${your_rds_instance_id_name}
#
#  Checks all RDS instances in a specific region
#  ./check-rds-events.rb --aws_region=${your_region}
#
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
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/awsclient"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sensu/sensu-aws/aws_session"
)

var (
	awsRegion    string
	dbInstanceId string
	ec2Client    *ec2.EC2
	rdsClient    *rds.RDS
)

func checkRdsEvents() {
	var success bool
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, ec2Client = awsclient.GetEC2Client(awsSession)
	if !success {
		return
	}
	resultRegions, err := ec2Client.DescribeRegions(nil)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	validRegion := false
	if resultRegions != nil && resultRegions.Regions != nil && len(resultRegions.Regions) > 0 {
		for _, region := range resultRegions.Regions {
			if *region.RegionName == awsRegion {
				validRegion = true
				break
			}
		}
	}
	if !validRegion {
		fmt.Println("CRITICAL : Invalid region specified!")
		return
	}
	success, rdsClient = awsclient.GetRDSClient(awsSession)
	if !success {
		return
	}
	clusters, err := getClusters()
	if err != nil || (!(clusters != nil && len(clusters) > 0)) {
		return
	}
	checkEvents(clusters)
}

func checkEvents(clusters []string) {
	criticalClusters := []string{}
	for _, cluster := range clusters {
		eventInput := &rds.DescribeEventsInput{}
		eventInput.SourceType = aws.String("db-instance")
		eventInput.SourceIdentifier = &cluster
		eventInput.StartTime = aws.Time(time.Now().Add(time.Duration(-24*60) * time.Minute))
		eventOutput, err := rdsClient.DescribeEvents(eventInput)

		if err != nil {
			fmt.Println("Error occurred while getting rds event details for db instance -", cluster, ",Error -", err)
			return
		}

		if eventOutput == nil || eventOutput.Events == nil || len(eventOutput.Events) == 0 {
			continue
		}

		for _, event := range eventOutput.Events {
			// we will need to filter out non-disruptive/basic operation events.
			//ie. the regular backup operations
			match, _ := regexp.MatchString("Backing up DB instance", *event.Message)
			if match {
				continue
			}
			match, _ = regexp.MatchString("Finished DB Instance backup", *event.Message)
			if match {
				continue
			}

			match, _ = regexp.MatchString("Restored from snapshot", *event.Message)
			if match {
				continue
			}

			match, _ = regexp.MatchString("DB instance created", *event.Message)
			if match {
				continue
			}
			// ie. Replication resumed
			match, _ = regexp.MatchString("Replication for the Read Replica resumed", *event.Message)
			if match {
				continue
			}

			// you can add more filters to skip more events.

			// draft the messages
			criticalClusters = append(criticalClusters, fmt.Sprintf("%s : %s \n", cluster, *event.Message))
		}
	}
	if len(criticalClusters) > 0 {
		fmt.Println("CRITICAL : Clusters w/ critical events :", criticalClusters)
	}
}

func getClusters() ([]string, error) {
	clusters := []string{}
	dbInstanceInput := &rds.DescribeDBInstancesInput{}
	filter := &rds.Filter{}
	if len(dbInstanceId) > 0 {
		filter.Name = aws.String("db-instance-id")
		filter.Values = []*string{aws.String(dbInstanceId)}
		dbInstanceInput.Filters = []*rds.Filter{filter}
	}
	dbClusterOutput, err := rdsClient.DescribeDBInstances(dbInstanceInput)
	if err != nil {
		fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
		return nil, err
	}

	if dbClusterOutput != nil && dbClusterOutput.DBInstances != nil && len(dbClusterOutput.DBInstances) > 0 {
		clusters = append(clusters, dbInstanceId)
	} else {
		fmt.Println("UNKNOWN :", dbInstanceId, "instance not found")
		return nil, nil
	}

	if len(dbInstanceId) == 0 {
		filter = &rds.Filter{}
		dbInstanceInput.Filters = []*rds.Filter{filter}
		dbClusterOutput, err = rdsClient.DescribeDBInstances(dbInstanceInput)
	}

	if err != nil {
		fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
		return nil, err
	}

	if dbClusterOutput != nil && dbClusterOutput.DBInstances != nil && len(dbClusterOutput.DBInstances) > 0 {
		for _, dbInstance := range dbClusterOutput.DBInstances {
			clusters = append(clusters, *dbInstance.DBInstanceIdentifier)
		}
	}
	return clusters, nil
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
	checkRdsEvents()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-rds-events",
		Short: "The Sensu Go Aws RDS handler for rds events management",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&awsRegion,
		"aws_region",
		"r",
		"us-east-1",
		"AWS Region")

	cmd.Flags().StringVarP(&dbInstanceId,
		"db_instance_id",
		"d",
		"",
		"DB instance identifier")
	return cmd
}
