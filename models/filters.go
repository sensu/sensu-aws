package models

import "github.com/aws/aws-sdk-go/service/ec2"

type Filters struct {
	Filters []*ec2.Filter
}
