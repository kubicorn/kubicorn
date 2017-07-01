package resources

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/kris-nova/kubicorn/logger"
)

type Asg struct {
	Resource
	Actual   *autoscaling.Group
	Expected *autoscaling.Group
}

func (r *Asg) Find() error {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{S(r.Name)},
	}
	output, err := r.AwsSdk.ASG.DescribeAutoScalingGroups(input)
	if err != nil {
		return err
	}
	lasg := len(output.AutoScalingGroups)
	if lasg == 1 {
		logger.Debug("Found %s [%s]", r.Type, r.Name)
		r.Actual = output.AutoScalingGroups[0]
	} else if lasg > 1 {
		logger.Warning("Found more than 1 %s for label [%s] %s", r.Type, r.Label, r.Name)
	}
	return nil
}

func (r *Asg) Apply() error {
	return nil
}

func (r *Asg) Init(known *cluster.Cluster, sdk cloud.Sdk) error {
	r.Type = "asg"
	r.Label = "kubicorn_asg_name"
	r.Name = known.Name
	r.Known = known
	r.AwsSdk = sdk.(*awsSdkGo.Sdk)
	logger.Debug("Loading AWS Resource [%s]", r.Type)
	return nil
}

func (r *Asg) RenderActual(actual *cluster.Cluster) error {
	if r.Actual == nil {
		return nil
	}

	// -----

	// Kris Left Off here

	// -----

	pool := &cluster.ServerPool{}
	actual.ServerPools = append(actual.ServerPools, pool)
	return nil
}
