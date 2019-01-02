package models

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type AwsInstance struct {
	Id         string
	LaunchTime time.Time
	Tags       []*ec2.Tag
}
