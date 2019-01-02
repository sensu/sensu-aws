package awsclient

/*
creates iam,ec2,elb,rds,sts, cloudwatch, s3, alb client
with valid awssession with roleArn support for rds and cloudwatch clients
*/

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
)

func newIAM(awsSession *session.Session) *iam.IAM {
	iamClient := iam.New(awsSession)
	return iamClient
}

func newEC2(awsSession *session.Session) *ec2.EC2 {
	ec2Client := ec2.New(awsSession)
	return ec2Client
}

func newCloudWatch(awsSession *session.Session) *cloudwatch.CloudWatch {
	cloudwatchClient := cloudwatch.New(awsSession)
	return cloudwatchClient
}

func newS3(awsSession *session.Session) *s3.S3 {
	s3Client := s3.New(awsSession)
	return s3Client
}

func newRDS(awsSession *session.Session) *rds.RDS {
	rdsClient := rds.New(awsSession)
	return rdsClient
}

func newELB(awsSession *session.Session) *elb.ELB {
	elbClient := elb.New(awsSession)
	return elbClient
}

func newELBV2(awsSession *session.Session) *elbv2.ELBV2 {
	elbv2Client := elbv2.New(awsSession)
	return elbv2Client
}

func newSTS(awsSession *session.Session) *sts.STS {
	stsClient := sts.New(awsSession)
	return stsClient
}

func GetElbClient(awsSession *session.Session) (bool, *elb.ELB) {
	var elbClient *elb.ELB
	if awsSession != nil {
		elbClient = newELB(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if elbClient == nil {
		fmt.Println("Error while getting elb client session")
		return false, nil
	}

	return true, elbClient
}

func GetElbV2Client(awsSession *session.Session) (bool, *elbv2.ELBV2) {
	var elbClient *elbv2.ELBV2
	if awsSession != nil {
		elbClient = newELBV2(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if elbClient == nil {
		fmt.Println("Error while getting elbv2 client session")
		return false, nil
	}

	return true, elbClient
}

func GetEC2Client(awsSession *session.Session) (bool, *ec2.EC2) {
	var ec2Client *ec2.EC2
	if awsSession != nil {
		ec2Client = newEC2(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if ec2Client == nil {
		fmt.Println("Error while getting ec2 client session")
		return false, nil
	}

	return true, ec2Client
}

func GetCloudWatchClient(awsSession *session.Session) (bool, *cloudwatch.CloudWatch) {
	var cloudWatClient *cloudwatch.CloudWatch
	if awsSession != nil {
		cloudWatClient = newCloudWatch(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if cloudWatClient == nil {
		fmt.Println("Error while getting cloudwatch client session")
		return false, nil
	}

	return true, cloudWatClient
}

func GetRDSClient(awsSession *session.Session) (bool, *rds.RDS) {
	var rdsClient *rds.RDS
	if awsSession != nil {
		rdsClient = newRDS(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if rdsClient == nil {
		fmt.Println("Error while getting rds client session")
		return false, nil
	}

	return true, rdsClient
}

func GetS3Client(awsSession *session.Session) (bool, *s3.S3) {
	var s3Client *s3.S3
	if awsSession != nil {
		s3Client = newS3(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if s3Client == nil {
		fmt.Println("Error while getting s3 client session")
		return false, nil
	}

	return true, s3Client
}

func getSTSClient(awsSession *session.Session) (bool, *sts.STS) {
	var stsClient *sts.STS
	if awsSession != nil {
		stsClient = newSTS(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if stsClient == nil {
		fmt.Println("Error while getting sts client session")
		return false, nil
	}

	return true, stsClient
}

func GetRDSClientWithRoleArn(awsSession *session.Session, roleArn string) (bool, *rds.RDS) {
	suceess, stscredentials := getStsCredentials(awsSession, roleArn)
	var rdsClient *rds.RDS
	if !suceess {
		return false, nil
	}
	if stscredentials != nil {
		provider := NewAssumeRoleCredentialsProvider(stscredentials)
		rdsClient = rds.New(awsSession,
			&aws.Config{Credentials: credentials.NewCredentials(provider)})
	}
	return false, rdsClient
}

func GetCloudWatchClientWithRoleArn(awsSession *session.Session, roleArn string) (bool, *cloudwatch.CloudWatch) {
	success, stscredentials := getStsCredentials(awsSession, roleArn)
	var cloudWatchClient *cloudwatch.CloudWatch
	if !success {
		return false, nil
	}
	if stscredentials != nil {
		provider := NewAssumeRoleCredentialsProvider(stscredentials)
		cloudWatchClient = cloudwatch.New(awsSession,
			&aws.Config{Credentials: credentials.NewCredentials(provider)})
	}
	return false, cloudWatchClient
}

func getStsCredentials(awsSession *session.Session, roleArn string) (bool, *sts.Credentials) {
	success, stsClient := getSTSClient(awsSession)
	if !success {
		return false, nil
	}
	roleInput := &sts.AssumeRoleInput{}
	roleInput.RoleArn = aws.String(roleArn)
	roleInput.RoleSessionName = aws.String(fmt.Sprintf("role@%v", time.Now().Unix()))
	roleOutput, err := stsClient.AssumeRole(roleInput)
	if err != nil {
		return false, nil
	}
	if roleOutput != nil {
		return true, roleOutput.Credentials
	}
	return false, nil
}

func NewAssumeRoleCredentialsProvider(credentials *sts.Credentials) *AssumeRoleCredentialsProvider {
	return &AssumeRoleCredentialsProvider{
		AssumeRoleCredentials: credentials,
	}
}

type AssumeRoleCredentialsProvider struct {
	AssumeRoleCredentials *sts.Credentials
}

func (c AssumeRoleCredentialsProvider) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     *c.AssumeRoleCredentials.AccessKeyId,
		SecretAccessKey: *c.AssumeRoleCredentials.SecretAccessKey,
		SessionToken:    *c.AssumeRoleCredentials.SessionToken,
		ProviderName:    "AssumeRoleCredentialsProvider",
	}, nil

}

func (c AssumeRoleCredentialsProvider) IsExpired() bool {
	return c.AssumeRoleCredentials.Expiration.After(time.Now())

}
