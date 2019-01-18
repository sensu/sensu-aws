package main

/*
#
# check-rds-pending
#
#
# DESCRIPTION:
#   This plugin checks rds clusters for pending maintenance action.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#  ./check-rds-pending --aws_region=${you_region}
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

	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/awsclient"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sensu/sensu-aws/aws_session"
)

var (
	awsRegion string
	rdsClient *rds.RDS
)

func checkRdsPending() {
	var success bool
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, rdsClient = awsclient.GetRDSClient(awsSession)
	if !success {
		return
	}
	clusters, err := getClusters()
	if err != nil || clusters == nil {
		return
	}
	checkPendingMaintenance(clusters)
}

func getClusters() ([]*string, error) {
	clusters := []*string{}
	dbInstanceInput := &rds.DescribeDBInstancesInput{}
	//fetch all clusters identifiers
	dbClusterOutput, err := rdsClient.DescribeDBInstances(dbInstanceInput)

	if err != nil {
		fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
		return nil, err
	}

	if dbClusterOutput != nil && dbClusterOutput.DBInstances != nil && len(dbClusterOutput.DBInstances) > 0 {
		for _, dbInstance := range dbClusterOutput.DBInstances {
			clusters = append(clusters, dbInstance.DBInstanceIdentifier)
		}
	}

	if !(clusters != nil && len(clusters) > 0) {
		fmt.Println("OK")
		return nil, nil
	}

	return clusters, nil
}

func checkPendingMaintenance(clusters []*string) {
	pendingMaintanceInput := &rds.DescribePendingMaintenanceActionsInput{}
	filter := &rds.Filter{}
	filter.Name = aws.String("db-instance-id")
	filter.Values = clusters
	pendingMaintanceInput.Filters = []*rds.Filter{filter}
	pendingMaintanceOutput, err := rdsClient.DescribePendingMaintenanceActions(pendingMaintanceInput)
	if err != nil {
		fmt.Println("Error :", err)
		return
	}

	if !(pendingMaintanceOutput != nil && pendingMaintanceOutput.PendingMaintenanceActions != nil && len(pendingMaintanceOutput.PendingMaintenanceActions) > 0) {
		return
	}

	fmt.Println("CRITICAL : Clusters w/ pending maintenance required:")
	for _, pendingMaintance := range pendingMaintanceOutput.PendingMaintenanceActions {
		fmt.Println(pendingMaintance.PendingMaintenanceActionDetails)
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
	checkRdsPending()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-rds-pending",
		Short: "The Sensu Go Aws RDS handler for rds maintenance management",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&awsRegion,
		"aws_region",
		"r",
		"us-east-1",
		"AWS Region")

	return cmd
}
