package resources
//
//import (
//	"fmt"
//	"github.com/aws/aws-sdk-go/service/autoscaling"
//	"github.com/kris-nova/kubicorn/apis/cluster"
//	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
//	"github.com/kris-nova/kubicorn/logger"
//)
//
//type Asg struct {
//	Resource
//	Actual               *Asg
//	Expected             *Asg
//	AssociatedServerPool *cluster.ServerPool
//}
//
//func (r *Asg) Parse() error {
//	actual := &Asg{}
//	expected := &Asg{}
//	var asg *autoscaling.Group
//	input := &autoscaling.DescribeAutoScalingGroupsInput{
//		AutoScalingGroupNames: []*string{S(r.AssociatedServerPool.Name)},
//	}
//	output, err := r.AwsSdk.ASG.DescribeAutoScalingGroups(input)
//	if err != nil {
//		return err
//	}
//	lasg := len(output.AutoScalingGroups)
//	if lasg > 0 {
//		if lasg > 1 {
//			return fmt.Errorf("More than 1 ASG found for name: %v", r.AssociatedServerPool.Name)
//		}
//		asg = output.AutoScalingGroups[0]
//		actual.ID = *asg.AutoScalingGroupName
//	}
//
//	r.Actual = actual
//	r.Expected = expected
//	return nil
//}
//
//func (r *Asg) Apply() error {
//	return nil
//}
//
//func (r *Asg) Init(known *cluster.Cluster, sdk *awsSdkGo.Sdk) error {
//	r.Type = "asg"
//	r.Label = r.Name
//	r.KnownCluster = known
//	r.AwsSdk = sdk
//	logger.Debug("Loading AWS Resource [%s]", r.Type)
//	return nil
//}
//
//func (r *Asg) Render() error {
//	return nil
//}
//
//func (r *Asg) Delete() error {
//	return nil
//}
