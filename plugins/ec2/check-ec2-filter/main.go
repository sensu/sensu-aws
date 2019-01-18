package main

/*
#
# check-ec2-filter
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
#   ./check-ec2-filter --filters="{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}"
#   ./check-ec2-filter --exclude_tags="{\"TAG_NAME\" : \"TAG_VALUE\"}" --compare=not
# NOTES:
#
# LICENSE:
#   Justin McCarty
#   Released under the same terms as Sensu (the MIT license); see LICENSE
#   for details.
#
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/awsclient"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sensu/sensu-aws/aws_session"
	"github.com/sensu/sensu-aws/models"
	"github.com/sensu/sensu-aws/utils"
)

var (
	ec2Client               *ec2.EC2
	criticalThreshold       int
	warningThreshold        int
	excludeTags             string
	compareValue            string
	detailedMessageRequired bool
	minRunningSecs          float64
	filters                 string
	awsRegion               string
)

func checkFilter() {
	var excludedTags map[string]*string
	var ec2Fileters models.Filters
	var awsInstances []models.AwsInstance
	var success bool
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, ec2Client = awsclient.GetEC2Client(awsSession)
	if !success {
		return
	}
	err := json.Unmarshal([]byte(excludeTags), &excludedTags)
	if err != nil {
		fmt.Println("Failed to unmarshal exclude tags details , ", err)
	}

	err = json.Unmarshal([]byte(filters), &ec2Fileters)
	if err != nil {
		fmt.Println("Failed to unmarshal filter data , ", err)
	}

	reservations, err := utils.GetReservations(ec2Client, ec2Fileters.Filters)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			tags := instance.Tags
			excludeIntance := false
			for _, tag := range tags {
				if excludedTags[*tag.Key] != nil && *excludedTags[*tag.Key] == *tag.Value {
					excludeIntance = true
				}
			}
			if !excludeIntance {
				timeDifference := time.Since(time.Now().Add(time.Duration(-10) * time.Minute)).Seconds()
				if !(timeDifference < minRunningSecs) {
					awsInstance := models.AwsInstance{Id: *instance.InstanceId, LaunchTime: *instance.LaunchTime, Tags: instance.Tags}
					awsInstances = append(awsInstances, awsInstance)
				}
			}
		}
	}

	selectedInstancesCount := len(awsInstances)
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Current Count : %d  ", selectedInstancesCount))
	if detailedMessageRequired && selectedInstancesCount > 0 {
		for _, awsInstance := range awsInstances {
			buffer.WriteString(fmt.Sprintf(", %s", awsInstance.Id))
		}
	}

	if compareValue == "equal" {
		if selectedInstancesCount == criticalThreshold {
			fmt.Println("Critical threshold for filter , ", buffer.String())
		}
		if selectedInstancesCount == warningThreshold {
			fmt.Println("Warning threshold for filter , ", buffer.String())
		}
	} else if compareValue == "not" {
		if selectedInstancesCount != criticalThreshold {
			fmt.Println("Critical threshold for filter , ", buffer.String())
		}
		if selectedInstancesCount != warningThreshold {
			fmt.Println("Warning threshold for filter , ", buffer.String())
		}
	} else if compareValue == "greater" {
		if selectedInstancesCount > criticalThreshold {
			fmt.Println("Critical threshold for filter , ", buffer.String())
		}
		if selectedInstancesCount > warningThreshold {
			fmt.Println("Warning threshold for filter , ", buffer.String())
		}
	} else if compareValue == "less" {
		if selectedInstancesCount < criticalThreshold {
			fmt.Println("Critical threshold for filter , ", buffer.String())
		}
		if selectedInstancesCount < warningThreshold {
			fmt.Println("Warning threshold for filter , ", buffer.String())
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
	checkFilter()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-ec2-filter",
		Short: "The Sensu Go Aws EC2 handler for instance filter management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region")
	cmd.Flags().IntVar(&criticalThreshold, "critical", 1, "Critical threshold for filter")
	cmd.Flags().IntVar(&warningThreshold, "warning", 2, "Warning threshold for filter',	")
	cmd.Flags().StringVar(&excludeTags, "exclude_tags", "{}", "JSON String Representation of tag values")
	cmd.Flags().StringVar(&compareValue, "compare", "equal", "Comparision operator for threshold: equal, not, greater, less")
	cmd.Flags().BoolVar(&detailedMessageRequired, "detailed_message", false, "Detailed description is required or not")
	cmd.Flags().Float64Var(&minRunningSecs, "min_running_secs", 0, "Minimum running seconds")
	cmd.Flags().StringVar(&filters, "filters", "{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}", "JSON String representation of Filters")

	return cmd
}
