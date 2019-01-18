package main

/*
#
# check-s3-tag
#
# DESCRIPTION:
#   This plugin checks if buckets have a set of tags.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-s3-tag --tag_keys=sensu
#
# LICENSE:
#   TODO
#
*/

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	"github.com/sensu/sensu-aws/aws_session"
	"github.com/sensu/sensu-aws/awsclient"
)

var (
	s3Client  *s3.S3
	tagKeys   string
	awsRegion string
)

func checkTag() {
	var success bool
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, s3Client = awsclient.GetS3Client(awsSession)
	if !success {
		return
	}

	tags := strings.Split(tagKeys, ",")
	tagMap := make(map[string]*string)
	missingTagsMap := make(map[string][]string)

	for _, tag := range tags {
		tagMap[tag] = &tag
	}

	input := &s3.ListBucketsInput{}
	output, err := s3Client.ListBuckets(input)
	if err != nil {
		fmt.Println(err)
		return
	}
	if output != nil && output.Buckets != nil && len(output.Buckets) > 1 {
		for _, bucket := range output.Buckets {
			bucketInput := &s3.GetBucketTaggingInput{Bucket: bucket.Name}
			bucketOutput, err := s3Client.GetBucketTagging(bucketInput)
			if err != nil {
				continue
			}
			if bucketOutput != nil && bucketOutput.TagSet != nil && len(bucketOutput.TagSet) > 0 {
				bucketTagMap := make(map[string]*string)
				for _, bucketTag := range bucketOutput.TagSet {
					bucketTagMap[*bucketTag.Key] = bucketTag.Key
				}
				for tag, tagValue := range tagMap {
					if bucketTagMap[tag] != nil {
						continue
					} else {
						missingTagsMap[*bucket.Name] = append(missingTagsMap[*bucket.Name], *tagValue)
					}
				}
			}
		}
	}

	if len(missingTagsMap) > 0 {
		for bucketName, tags := range missingTagsMap {
			fmt.Println("CRITICAL : Missing tags for bucket", bucketName, ":", tags)
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
	checkTag()
	return nil
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-s3-tag",
		Short: "The Sensu Go Aws Bucket handler for tag management",
		RunE:  run,
	}

	flag.StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	flag.StringVar(&tagKeys, "tag_keys", "", "Comma seperated Tag Keys")

	cmd.Flags().StringVarP(&awsRegion,
		"aws_region",
		"r",
		"us-east-1",
		"AWS Region")

	cmd.Flags().StringVarP(&tagKeys,
		"tag_keys",
		"t",
		"",
		"Comma seperated Tag Keys")

	_ = cmd.MarkFlagRequired("tag_keys")
	return cmd
}
