package main

/*
#
# check-s3-bucket
#
# DESCRIPTION:
#   This plugin checks a bucket and alerts if not exists
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./check-s3-bucket --bucket_name=mybucket
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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/aws_session"
	"github.com/sensu/sensu-aws/awsclient"
)

var (
	s3Client   *s3.S3
	awsRegion  string
	bucketName string
)

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
	checkS3Bucket()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-s3-bucket",
		Short: "The Sensu Go Aws Bucket handler for bucket management",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&awsRegion,
		"aws_region",
		"r",
		"us-east-1",
		"AWS Region")

	cmd.Flags().StringVarP(&bucketName,
		"bucket_name",
		"b",
		"",
		"An S3 bucket to check")

	_ = cmd.MarkFlagRequired("bucket_name")
	return cmd
}

func checkS3Bucket() {
	var success bool
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, s3Client = awsclient.GetS3Client(awsSession)
	if !success {
		return
	}
	input := &s3.HeadBucketInput{Bucket: aws.String(bucketName)}
	_, err := s3Client.HeadBucket(input)
	if err != nil && err.(awserr.Error).Code() == "NotFound" {
		fmt.Println("CRITICAL:", bucketName, "bucket not found")
	} else if err != nil {
		fmt.Println("CRITICAL:", bucketName, "-", err.(awserr.Error).Message())
	} else {
		fmt.Println("OK:", bucketName, "bucket found")
	}
}
