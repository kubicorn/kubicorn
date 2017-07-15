package awsSdkGo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Sdk struct {
	Ec2 *ec2.EC2
	S3  *s3.S3
	ASG *autoscaling.AutoScaling
}

func NewSdk(region string) (*Sdk, error) {
	sdk := &Sdk{}
	session, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}
	sdk.Ec2 = ec2.New(session)
	sdk.ASG = autoscaling.New(session)
	sdk.S3 = s3.New(session)
	return sdk, nil
}
