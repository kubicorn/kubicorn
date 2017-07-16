package resources

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/logger"
)

type Asg struct {
	Shared
	MinCount     int
	MaxCount     int
	InstanceType string
	Image        string
	ServerPool   *cluster.ServerPool
}

func (r *Asg) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("asg.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached ASG [actual]")
		return r.CachedActual, nil
	}
	actual := &Asg{
		Shared: Shared{
			Name:        r.Name,
			Tags:        make(map[string]string),
			TagResource: r.TagResource,
		},
		ServerPool: r.ServerPool,
	}

	if r.ServerPool.Identifier != "" {
		input := &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{S(r.ServerPool.Identifier)},
		}
		output, err := Sdk.ASG.DescribeAutoScalingGroups(input)
		if err != nil {
			return nil, err
		}
		lasg := len(output.AutoScalingGroups)
		if lasg != 1 {
			return nil, fmt.Errorf("Found [%d] ASGs for ID [%s]", lasg, r.ServerPool.Identifier)
		}
		asg := output.AutoScalingGroups[0]
		for _, tag := range asg.Tags {
			key := *tag.Key
			val := *tag.Value
			actual.Tags[key] = val
		}
		actual.MaxCount = int(*asg.MaxSize)
		actual.MinCount = int(*asg.MinSize)
		actual.CloudID = *asg.AutoScalingGroupName
		actual.Name = *asg.AutoScalingGroupName
	}
	r.CachedActual = actual
	return actual, nil
}
func (r *Asg) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("asg.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using cached ASG [expected]")
		return r.CachedExpected, nil
	}
	expected := &Asg{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": known.Name,
			},
			CloudID:     known.Network.Identifier,
			Name:        r.Name,
			TagResource: r.TagResource,
		},
		ServerPool: r.ServerPool,
		MaxCount:   r.ServerPool.MaxCount,
		MinCount:   r.ServerPool.MinCount,
	}
	r.CachedExpected = expected
	return expected, nil
}
func (r *Asg) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("asg.Apply")
	applyResource := expected.(*Asg)
	isEqual, err := compare.IsEqual(actual.(*Asg), expected.(*Asg))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}
	subnetId := ""
	for _, sp := range applyCluster.ServerPools {
		if sp.Name == r.Name {
			for _, sn := range sp.Subnets {
				if sn.Name == r.Name {
					subnetId = sn.Identifier
				}
			}
		}
	}
	if subnetId == "" {
		return nil, fmt.Errorf("Unable to find subnet id")
	}

	newResource := &Asg{}
	input := &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName:    &r.Name,
		MinSize:                 I64(expected.(*Asg).MinCount),
		MaxSize:                 I64(expected.(*Asg).MaxCount),
		LaunchConfigurationName: &r.Name,
		VPCZoneIdentifier:       &subnetId,
	}
	_, err = Sdk.ASG.CreateAutoScalingGroup(input)
	if err != nil {
		return nil, err
	}
	logger.Info("Created Asg [%s]", r.Name)

	// Todo popualte newResource here with values from API
	newResource.Name = r.Name
	newResource.CloudID = r.Name
	newResource.MaxCount = r.MaxCount
	newResource.MinCount = r.MinCount

	err = newResource.Tag(applyResource.Tags)
	if err != nil {
		return nil, fmt.Errorf("Unable to tag new VPC: %v", err)
	}
	return newResource, nil
}
func (r *Asg) Delete(actual cloud.Resource, known *cluster.Cluster) error {
	logger.Debug("asg.Delete")
	deleteResource := actual.(*Asg)
	if deleteResource.CloudID == "" {
		return fmt.Errorf("Unable to delete ASG resource without ID [%s]", deleteResource.Name)
	}
	// Delete ASG API

	input := &autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: &actual.(*Asg).CloudID,
		ForceDelete:          B(true),
	}
	_, err := Sdk.ASG.DeleteAutoScalingGroup(input)
	if err != nil {
		return err
	}
	logger.Info("Deleted ASG [%s]", actual.(*Asg).CloudID)
	return nil
}

func (r *Asg) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("asg.Render")
	serverPool := &cluster.ServerPool{}
	serverPool.MaxCount = renderResource.(*Asg).MaxCount
	serverPool.MaxCount = renderResource.(*Asg).MinCount
	serverPool.Name = renderResource.(*Asg).Name
	serverPool.Identifier = renderResource.(*Asg).CloudID

	found := false
	for i := 0; i < len(renderCluster.ServerPools); i++ {
		if renderCluster.ServerPools[i].Name == renderResource.(*Asg).Name {
			renderCluster.ServerPools[i].MaxCount = renderResource.(*Asg).MaxCount
			renderCluster.ServerPools[i].MinCount = renderResource.(*Asg).MinCount
			renderCluster.ServerPools[i].Name = renderResource.(*Asg).Name
			renderCluster.ServerPools[i].Identifier = renderResource.(*Asg).CloudID
			found = true
		}
	}
	if !found {
		renderCluster.ServerPools = append(renderCluster.ServerPools, serverPool)
	}

	return renderCluster, nil
}

// kris left off here
// 2017-07-05T00:22:36-06:00 [✖]  Unable to reconcile cluster: Unable to tag new VPC: ValidationError: Incomplete tags information for these tags. Tag - 'Key:Name,PropagateAtLaunch:null,ResourceId:knova-amazon-master,ResourceType:auto-scaling-group' , Tag - 'Key:KubernetesCluster,PropagateAtLaunch:null,ResourceId:knova-amazon-master,ResourceType:auto-scaling-group' ,
// status code: 400, request id: 569ec7d4-614a-11e7-a82d-f902128f85b7

func (r *Asg) Tag(tags map[string]string) error {
	logger.Debug("asg.Tag")
	tagInput := &autoscaling.CreateOrUpdateTagsInput{}
	for key, val := range tags {
		logger.Debug("Registering Asg tag [%s] %s", key, val)
		tagInput.Tags = append(tagInput.Tags, &autoscaling.Tag{
			Key:               S("%s", key),
			Value:             S("%s", val),
			ResourceType:      S("auto-scaling-group"),
			ResourceId:        &r.CloudID,
			PropagateAtLaunch: B(true),
		})
	}
	_, err := Sdk.ASG.CreateOrUpdateTags(tagInput)
	if err != nil {
		return err
	}
	return nil
}
