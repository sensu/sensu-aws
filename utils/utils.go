package utils

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetReservations(ec2Client *ec2.EC2, filters []*ec2.Filter) ([]*ec2.Reservation, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	result, err := ec2Client.DescribeInstances(input)
	if err != nil {
		return nil, err
	}

	return result.Reservations, nil
}
