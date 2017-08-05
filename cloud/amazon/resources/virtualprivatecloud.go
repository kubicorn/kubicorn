// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/logger"
)

type Vpc struct {
	Shared
	CIDR string
}

func (r *Vpc) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("vpc.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached VPC [actual]")
		return r.CachedActual, nil
	}
	actual := &Vpc{
		Shared: Shared{
			Tags: make(map[string]string),
		},
	}
	if known.Network.Identifier != "" {
		input := &ec2.DescribeVpcsInput{
			VpcIds: []*string{&known.Network.Identifier},
		}
		output, err := Sdk.Ec2.DescribeVpcs(input)
		if err != nil {
			return nil, err
		}
		lvpc := len(output.Vpcs)
		if lvpc != 1 {
			return nil, fmt.Errorf("Found [%d] VPCs for ID [%s]", lvpc, known.Network.Identifier)
		}
		actual.CloudID = *output.Vpcs[0].VpcId
		actual.CIDR = *output.Vpcs[0].CidrBlock
		for _, tag := range output.Vpcs[0].Tags {
			key := *tag.Key
			val := *tag.Value
			actual.Tags[key] = val
		}
	}
	r.CachedActual = actual
	return actual, nil
}
func (r *Vpc) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("vpc.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using cached VPC [expected]")
		return r.CachedExpected, nil
	}
	expected := &Vpc{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": known.Name,
			},
			CloudID: known.Network.Identifier,
			Name:    r.Name,
		},
		CIDR: known.Network.CIDR,
	}
	r.CachedExpected = expected
	return expected, nil
}
func (r *Vpc) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("vpc.Apply")
	applyResource := expected.(*Vpc)
	isEqual, err := compare.IsEqual(actual.(*Vpc), expected.(*Vpc))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}
	newResource := &Vpc{}
	input := &ec2.CreateVpcInput{
		CidrBlock: &applyResource.CIDR,
	}
	output, err := Sdk.Ec2.CreateVpc(input)
	if err != nil {
		return nil, fmt.Errorf("Unable to create new VPC: %v", err)
	}

	minput1 := &ec2.ModifyVpcAttributeInput{
		EnableDnsHostnames: &ec2.AttributeBooleanValue{
			Value: B(true),
		},
		VpcId: output.Vpc.VpcId,
	}
	_, err = Sdk.Ec2.ModifyVpcAttribute(minput1)
	if err != nil {
		return nil, err
	}

	minput2 := &ec2.ModifyVpcAttributeInput{
		EnableDnsSupport: &ec2.AttributeBooleanValue{
			Value: B(true),
		},
		VpcId: output.Vpc.VpcId,
	}
	_, err = Sdk.Ec2.ModifyVpcAttribute(minput2)
	if err != nil {
		return nil, err
	}

	logger.Info("Created VPC [%s]", *output.Vpc.VpcId)
	newResource.CIDR = *output.Vpc.CidrBlock
	newResource.CloudID = *output.Vpc.VpcId
	err = newResource.Tag(applyResource.Tags)
	if err != nil {
		return nil, fmt.Errorf("Unable to tag new VPC: %v", err)
	}
	newResource.Name = applyResource.Name
	return newResource, nil
}
func (r *Vpc) Delete(actual cloud.Resource, known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("vpc.Delete")
	deleteResource := actual.(*Vpc)
	if deleteResource.CloudID == "" {
		return nil, fmt.Errorf("Unable to delete VPC resource without ID [%s]", deleteResource.Name)
	}
	input := &ec2.DeleteVpcInput{
		VpcId: &actual.(*Vpc).CloudID,
	}
	_, err := Sdk.Ec2.DeleteVpc(input)
	if err != nil {
		return nil, err
	}
	logger.Info("Deleted VPC [%s]", actual.(*Vpc).CloudID)

	newResource := &Vpc{}
	newResource.Name = actual.(*Vpc).Name
	newResource.Tags = actual.(*Vpc).Tags

	return newResource, nil
}

func (r *Vpc) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("vpc.Render")
	renderCluster.Network.CIDR = renderResource.(*Vpc).CIDR
	renderCluster.Network.Identifier = renderResource.(*Vpc).CloudID
	renderCluster.Network.Name = renderResource.(*Vpc).Name
	return renderCluster, nil
}

func (r *Vpc) Tag(tags map[string]string) error {
	logger.Debug("vpc.Tag")
	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.CloudID},
	}
	for key, val := range tags {
		logger.Debug("Registering Vpc tag [%s] %s", key, val)
		tagInput.Tags = append(tagInput.Tags, &ec2.Tag{
			Key:   S("%s", key),
			Value: S("%s", val),
		})
	}
	_, err := Sdk.Ec2.CreateTags(tagInput)
	if err != nil {
		return err
	}
	return nil
}
