package main

/*
#
# check-s3-object
#
# DESCRIPTION:
#   This plugin checks if a file exists in a bucket and/or is not too old.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./check-s3-object.go --bucket_name=sreejita-testing --key_prefix=s3
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
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sreejita-biswas/aws-handler/awsclient"
	"github.com/sreejita-biswas/aws-handler/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sreejita-biswas/aws-handler/aws_session"
)

var (
	s3Client                *s3.S3
	filters                 string
	useIamRole              bool
	bucketName              string
	allBuckets              bool
	excludeBuckets          string
	keyName                 string
	keyPrefix               string
	warningAge              float64
	criticalAge             float64
	okZeroSize              bool
	warningSize             int64
	criticalSize            int64
	compareSize             string
	noCritOnMultipleObjects bool
	awsRegion               string
)

func checkObject() {
	var age time.Duration
	var size int64
	var keyFullName string
	var success bool

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, s3Client = awsclient.GetS3Client(awsSession)
	if !success {
		return
	}

	if (len(strings.TrimSpace(keyName)) == 0 && len(strings.TrimSpace(keyPrefix)) == 0) || (len(strings.TrimSpace(keyName)) > 0 && len(strings.TrimSpace(keyPrefix)) > 0) {
		fmt.Println("Need one option between \"key_name\" and \"key_prefix\"")
		return
	}

	if len(strings.TrimSpace(keyName)) > 0 {
		input := &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(keyName)}
		output, err := s3Client.HeadObject(input)
		if err != nil {
			printErroMessage(err, keyName)
			return
		}
		if output != nil {
			age = time.Since(*output.LastModified)
			size = *output.ContentLength
			keyFullName = keyName
			printMesaage(age, keyFullName, size)
		}
	} else if len(strings.TrimSpace(keyPrefix)) > 0 {
		input := &s3.ListObjectsInput{Bucket: aws.String(bucketName), Prefix: aws.String(keyPrefix)}
		output, err := s3Client.ListObjects(input)
		if err != nil {
			printErroMessage(err, keyFullName)
			return
		}
		if output == nil || output.Contents == nil || len(output.Contents) < 1 {
			fmt.Println(fmt.Sprintf("CRITICAL : Object with prefix \"%s\" not found in bucket '%s'", keyPrefix, bucketName))
			return
		}

		if output != nil || output.Contents != nil || len(output.Contents) > 1 {
			if !noCritOnMultipleObjects {
				fmt.Println(fmt.Sprintf("CRITICAL : Your prefix \"%s\" return too much files, you need to be more specific", keyPrefix))
				return
			}
			utils.SortContents(output.Contents)
		}

		keyFullName = *output.Contents[0].Key
		age = time.Since(*output.Contents[0].LastModified)
		size = *output.Contents[0].Size
		printMesaage(age, keyFullName, size)
	}
}

func checkAge(age time.Duration, keyName string) {
	if age.Seconds() > criticalAge {
		fmt.Println(fmt.Sprintf("CRITICAL : S3 Object '%s' size : '%d' octets (Bucket - '%s')", keyName, age, bucketName))
		return
	}
	if age.Seconds() > warningAge {
		fmt.Println(fmt.Sprintf("WARNING : S3 Object '%s' size : '%d' octets (Bucket - '%s')", keyName, age, bucketName))
		return
	}
	fmt.Println(fmt.Sprintf("OK : S3 Object '%s' exists in bucket '%s'", keyName, bucketName))
}

func checkSize(size int64, keyName string) {
	if compareSize == "not" {
		if size != criticalSize {
			fmt.Println(fmt.Sprintf("CRITICAL : S3 Object '%s' size : '%d' octets (Bucket - '%s')", keyName, size, bucketName))
			return
		}
		if size != warningSize {
			fmt.Println(fmt.Sprintf("WARNING : S3 Object '%s' size : '%d' octets (Bucket - '%s')", keyName, size, bucketName))
			return
		}
	}

	if compareSize == "greater" {
		if size > criticalSize {
			fmt.Println(fmt.Sprintf("CRITICAL : S3 Object '%s' size : '%d' octets (Bucket - '%s')", keyName, size, bucketName))
			return
		}
		if size > warningSize {
			fmt.Println(fmt.Sprintf("WARNING : S3 Object '%s' size : '%d' octets (Bucket - '%s')", keyName, size, bucketName))
			return
		}

		fmt.Println(fmt.Sprintf("OK : S3 Object '%s' exists in bucket '%s'", keyName, bucketName))
	}

	if compareSize == "less" {
		if size < criticalSize {
			fmt.Println(fmt.Sprintf("CRITICAL : S3 Object '%s' size : '%d' octets (Bucket - '%s')", keyName, size, bucketName))
			return
		}
		if size < warningSize {
			fmt.Println(fmt.Sprintf("WARNING : S3 Object '%s' size : '%d' octets (Bucket - '%s')", keyName, size, bucketName))
			return
		}
		fmt.Println(fmt.Sprintf("OK : S3 Object '%s' exists in bucket '%s'", keyName, bucketName))
	}
}

func printErroMessage(err error, keyFullName string) {
	if err.(awserr.Error).Code() == "NotFound" {
		fmt.Println(fmt.Sprintf("CRITICAL : S3 Object '%s' not found in bucket - '%s'", keyFullName, bucketName))
	} else {
		fmt.Println(err.(awserr.Error).Message())
	}
}

func printMesaage(age time.Duration, keyFullName string, size int64) {
	checkAge(age, keyFullName)
	if size != 0 {
		checkSize(size, keyFullName)
	} else if !okZeroSize {
		fmt.Println(fmt.Sprintf("CRITICAL : S3 Object '%s' is empty (Bucket - '%s')", keyFullName, bucketName))
	}
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-s3-object",
		Short: "The Sensu Go Aws S3 Object handler for object management",
		RunE:  run,
	}
	cmd.Flags().StringVarP(&awsRegion,
		"aws_region",
		"r",
		"us-east-1",
		"AWS Region (defaults to us-east-1).")
	cmd.Flags().BoolVarP(&useIamRole,
		"use_iam_role",
		"u",
		false,
		"Use IAM role authenticiation. Instance must have IAM role assigned for this to work")
	cmd.Flags().StringVarP(&bucketName,
		"bucket_name",
		"b",
		"",
		"The name of the S3 bucket where object lives")
	cmd.Flags().StringVarP(&keyName,
		"key_name",
		"k",
		"",
		"The name of key in the bucket")
	cmd.Flags().StringVarP(&keyPrefix,
		"key_prefix",
		"p",
		"",
		"Prefix key to search on the bucket")
	cmd.Flags().Float64VarP(&warningAge,
		"warning_age",
		"w",
		90000,
		"Warn if mtime greater than provided age in seconds")
	cmd.Flags().Float64VarP(&criticalAge,
		"critical_age",
		"c",
		126000,
		"Critical if mtime greater than provided age in seconds")
	cmd.Flags().BoolVarP(&okZeroSize,
		"ok_zero_size",
		"z",
		true,
		"OK if file has zero size'")
	cmd.Flags().Int64VarP(&warningSize,
		"warning_size",
		"",
		0,
		"Warning threshold for size")
	cmd.Flags().Int64VarP(&criticalSize,
		"critical_size",
		"",
		0,
		"Critical threshold for size")
	cmd.Flags().StringVarP(&compareSize,
		"operator-size",
		"",
		"equal",
		"Comparision operator for threshold: equal, not, greater, less")
	cmd.Flags().BoolVarP(&noCritOnMultipleObjects,
		"no_crit_on_multiple_objects",
		"",
		true,
		"If this flag is set, sort all matching objects by last_modified date and check against the newest. By default, this check will return a CRITICAL result if multiple matching objects are found.")
	_ = cmd.MarkFlagRequired("bucket_name")
	return cmd
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
	checkObject()
	return nil
}
