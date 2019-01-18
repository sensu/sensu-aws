package main

/*
#
# check-s3-bucket-visibility
#
# DESCRIPTION:
#   This plugin checks a bucket for website configuration and bucket policy.
#   It alerts if the bucket has a website configuration, or a policy that has
#   Get or List actions.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./check-s3-bucket-visibility.go --exclude_buckets_regx=sensu --bucket_names=ssensu-ec2,sensu-ec3 --exclude_cuckets=sensu-ec3
#
# NOTES:
#
# LICENSE:
#  TODO
#
*/

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/awsclient"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sensu/sensu-aws/aws_session"
)

var (
	s3Client           *s3.S3
	filters            string
	awsRegion          string
	bucketNames        string
	allBuckets         bool
	excludeBuckets     string
	excludeBucketsRegx string
	criticalOnMissing  bool
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
	checkBucketVisibility()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-s3-bucket-visibility",
		Short: "The Sensu Go Aws Bucket handler for bucket visibility management",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&awsRegion,
		"aws_region",
		"r",
		"us-east-1",
		"AWS Region")

	cmd.Flags().StringVarP(&bucketNames,
		"bucket_names",
		"b",
		"",
		"A comma seperated list of S3 buckets to check")

	cmd.Flags().BoolVarP(&allBuckets,
		"all_buckets",
		"a",
		false,
		"If all buckets are true it will look at any buckets that we have access to in the region")

	cmd.Flags().StringVarP(&excludeBuckets,
		"exclude_buckets",
		"x",
		"",
		"A comma seperated list of buckets to ignore that are expected to have loose permissions")

	cmd.Flags().StringVarP(&excludeBucketsRegx,
		"exclude_buckets_regx",
		"r",
		"",
		"A regex to filter out bucket names")

	cmd.Flags().BoolVarP(&criticalOnMissing,
		"critical_on_missing",
		"m",
		false,
		"The check will fail with CRITICAL rather than WARN when a bucket is not found")

	_ = cmd.MarkFlagRequired("bucket_names")
	return cmd
}

func checkBucketVisibility() {
	var bucketsTobeExcluded []string
	var excludeBucket bool
	var success bool
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, s3Client = awsclient.GetS3Client(awsSession)
	if !success {
		return
	}
	if len(strings.TrimSpace(excludeBuckets)) > 0 {
		bucketsTobeExcluded = strings.Split(excludeBuckets, ",")
	}
	excludeBucketsMap := make(map[string]*string)
	for _, bucket := range bucketsTobeExcluded {
		excludeBucketsMap[bucket] = &bucket
	}
	buckets := strings.Split(bucketNames, ",")
	for _, bucket := range buckets {
		excludeBucket = false
		if len(strings.TrimSpace(excludeBucketsRegx)) > 0 {
			excludeBucket, _ = regexp.MatchString(excludeBucketsRegx, bucket)
		}
		if excludeBucket || excludeBucketsMap[bucket] != nil {
			excludeBucket = true
		} else {
			excludeBucket = false
		}
		if !excludeBucket {
			input := &s3.GetBucketWebsiteInput{Bucket: aws.String(bucket)}
			_, err := s3Client.GetBucketWebsite(input)
			if err != nil {
				if err.(awserr.Error).Code() == "NoSuchBucket" {
					if criticalOnMissing {
						fmt.Println(fmt.Sprintf("CRITICAL:'%s' bucket does not exist", bucket))

					} else {
						fmt.Println(fmt.Sprintf("WARNING:'%s' bucket does not exist", bucket))
					}
					continue
				}
				if err.(awserr.Error).Code() == "NoSuchWebsiteConfiguration" {
					fmt.Println(fmt.Sprintf("OK:'%s' bucket does not have a website configuration", bucket))
				}
			} else {
				fmt.Println(fmt.Sprintf("CRITICAL:'%s' bucket website configuration found", bucket))
			}
			policyInput := &s3.GetBucketPolicyInput{Bucket: aws.String(bucket)}
			policyResponse, err := s3Client.GetBucketPolicy(policyInput)
			if err != nil {
				if err.(awserr.Error).Code() == "NoSuchBucket" {
					if criticalOnMissing {
						fmt.Println(fmt.Sprintf("CRITICAL:'%s' bucket does not exist", bucket))
					} else {
						fmt.Println(fmt.Sprintf("WARNING:'%s' bucket does not exist", bucket))
					}
				}
				if err.(awserr.Error).Code() == "NoSuchBucketPolicy" {
					fmt.Println(fmt.Sprintf("OK:'%s' bucket policy does not exist", bucket))
				}
			} else if policyResponse != nil {
				fmt.Println(fmt.Sprintf("CRITICAL:'%s' bucket policy too permissive", bucket))
			}
		}
	}
}
