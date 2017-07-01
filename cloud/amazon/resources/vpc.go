package resources

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/kris-nova/kubicorn/logger"
)

type Vpc struct {
	Resource
	Actual   *ec2.Vpc
	Expected *ec2.Vpc
}

func (r *Vpc) Find() error {
	input := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   S("tag:%s", r.Label),
				Values: []*string{S(r.Name)},
			},
		},
	}
	output, err := r.AwsSdk.Ec2.DescribeVpcs(input)
	if err != nil {
		return err
	}
	lvpc := len(output.Vpcs)
	if lvpc == 1 {
		logger.Debug("Found %s [%s]", r.Type, r.Name)
		r.Actual = output.Vpcs[0]
	} else if lvpc > 1 {
		logger.Warning("Found more than 1 %s for label [%s] %s", r.Type, r.Label, r.Name)
	}
	return nil
}

func (r *Vpc) Apply() error {
	return nil
}

func (r *Vpc) Init(known *cluster.Cluster, sdk cloud.Sdk) error {
	r.Type = "vpc"
	r.Label = "kubicorn_vpc_name"
	r.Name = known.Name
	r.Known = known
	r.AwsSdk = sdk.(*awsSdkGo.Sdk)
	logger.Debug("Loading AWS Resource [%s]", r.Type)
	return nil
}

func (r *Vpc) RenderActual(actual *cluster.Cluster) error {
	if r.Actual == nil {
		return nil
	}
	actual.NetworkIdentifier = *r.Actual.VpcId
	return nil
}
