package main

/*
#
# check-ebs-snapshots
#
# DESCRIPTION:
#   Check EC2 Attached Volumes for Snapshots.  Only for Volumes with a Name tag.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-ebs-snapshots --check_ignored=false
#
# NOTES:
#   When using check_ignored flag value as true, any volume that has a tag-key of "IGNORE_BACKUP" will
#   be ignored.
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
	checkIgnored      bool
	period            int64
)

func checkSnapshot() {
	var errors []string
	var success bool

	volumeInput := &ec2.DescribeVolumesInput{}
	tagNames := []string{}

	filter := &ec2.Filter{}
	filter.Name = aws.String("attachment.status")
	filter.Values = []*string{aws.String("attached")}
	volumeInput.Filters = []*ec2.Filter{filter}
	filter2 := &ec2.Filter{}
	filter2.Name = aws.String("tag-key")
	filter2.Values = []*string{aws.String("Name")}

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

	if volumes != nil {
		for _, volume := range volumes.Volumes {
			tags := volume.Tags
			ignoreVolume := false
			tagNames = []string{}
			if volume.Tags != nil && len(volume.Tags) > 0 {
				for _, tag := range tags {
					if checkIgnored && *tag.Key == "IGNORE_BACKUP" {
						ignoreVolume = true
						break
					} else {
						tagNames = append(tagNames, *tag.Key)
					}
				}
				if ignoreVolume {
					continue
				}
				latestSnapshot, err := getLatestSnapshot(*volume.VolumeId)
				if err != nil {
					fmt.Println("Error : ", err)
					return
				}
				if latestSnapshot != nil {
					timeDiffrence := aws.Time(time.Now().Add(-time.Duration(period*24*60) * time.Minute)).Sub(*latestSnapshot.StartTime)
					if timeDiffrence.Seconds() < 0 {
						errors = append(errors, fmt.Sprintf("%v latest snapshot is %v for Voulme %s \n", tagNames, *latestSnapshot.StartTime, *volume.VolumeId))
					}
				} else {
					errors = append(errors, fmt.Sprintf("%v has no snapshot for Voulme %s \n", tagNames, *volume.VolumeId))
				}
			}
		}

		if len(errors) > 0 {
			fmt.Println("Warning : ", errors)
		} else {
			fmt.Println("Ok")
		}
	}
}

func getLatestSnapshot(volumeId string) (*ec2.Snapshot, error) {
	var latestSnapshot *ec2.Snapshot
	filter := &ec2.Filter{}
	filter.Name = aws.String("volume-id")
	filter.Values = []*string{&volumeId}
	snapshotInput := &ec2.DescribeSnapshotsInput{}
	snapshotInput.Filters = []*ec2.Filter{filter}
	snapshots, err := ec2Client.DescribeSnapshots(snapshotInput)
	if err != nil {
		return nil, err
	}
	if snapshots != nil && snapshots.Snapshots != nil && len(snapshots.Snapshots) >= 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		minimumTimeDifference = -1
		for _, snapshot := range snapshots.Snapshots {
			timeDifference = time.Since(*snapshot.StartTime).Seconds()
			if minimumTimeDifference == -1 || timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				latestSnapshot = snapshot
			}
		}
		return latestSnapshot, nil

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
	checkSnapshot()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-ebs-snapshot",
		Short: "The Sensu Go Aws EBS handler for snapshot management",
		RunE:  run,
	}

	cmd.Flags().StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	cmd.Flags().BoolVar(&checkIgnored, "check_ignored", true, "mark as true to ignore volumes with an IGNORE_BACKUP tag")
	cmd.Flags().Int64Var(&period, "period", 7, "Length in time to alert on missing snapshots")
	return cmd
}
