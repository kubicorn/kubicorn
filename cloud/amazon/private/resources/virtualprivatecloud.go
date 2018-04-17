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
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &Vpc{}

type Vpc struct {
	Shared
	CIDR string
}

func (r *Vpc) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("vpc.Actual")

	newResource := &Vpc{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
		CIDR: immutable.ProviderConfig().Network.CIDR,
	}

	if immutable.ProviderConfig().Network.Identifier != "" {
		output, err := Sdk.Ec2.DescribeVpcs(&ec2.DescribeVpcsInput{
			VpcIds: []*string{&immutable.ProviderConfig().Network.Identifier},
		})
		if err != nil {
			return nil, nil, err
		}
		vpc := output.Vpcs[0]

		newResource.Identifier = *vpc.VpcId
		for _, tag := range vpc.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Vpc) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("vpc.Expected")

	newResource := &Vpc{
		Shared: Shared{
			Identifier: immutable.ProviderConfig().Network.Identifier,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
			},
		},
		CIDR: immutable.ProviderConfig().Network.CIDR,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Vpc) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("vpc.Apply")

	applyResource := expected.(*Vpc)
	isEqual, err := compare.IsEqual(actual.(*Vpc), applyResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	output, err := Sdk.Ec2.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: &applyResource.CIDR,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new VPC: %v", err)
	}
	vpcID := output.Vpc.VpcId

	logger.Info("Waiting for VPC [%s] to be available", *vpcID)
	err = Sdk.Ec2.WaitUntilVpcAvailable(&ec2.DescribeVpcsInput{
		VpcIds: []*string{vpcID},
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = Sdk.Ec2.ModifyVpcAttribute(&ec2.ModifyVpcAttributeInput{
		EnableDnsHostnames: &ec2.AttributeBooleanValue{
			Value: B(true),
		},
		VpcId: vpcID,
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = Sdk.Ec2.ModifyVpcAttribute(&ec2.ModifyVpcAttributeInput{
		EnableDnsSupport: &ec2.AttributeBooleanValue{
			Value: B(true),
		},
		VpcId: vpcID,
	})
	if err != nil {
		return nil, nil, err
	}

	logger.Success("Created VPC [%s]", *vpcID)

	newResource := &Vpc{
		Shared: Shared{
			Identifier: *vpcID,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
		CIDR: *output.Vpc.CidrBlock,
	}
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new VPC: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Vpc) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("vpc.Delete")

	deleteResource := actual.(*Vpc)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete VPC resource without ID [%s]", deleteResource.Name)
	}

	_, err := Sdk.Ec2.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: &deleteResource.Identifier,
	})
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted VPC [%s]", deleteResource.Identifier)

	newResource := &Vpc{
		Shared: Shared{
			Name: deleteResource.Name,
			Tags: deleteResource.Tags,
		},
		CIDR: deleteResource.CIDR,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Vpc) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("vpc.Render")

	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.Network.CIDR = newResource.(*Vpc).CIDR
	providerConfig.Network.Identifier = newResource.(*Vpc).Identifier
	providerConfig.Network.Name = newResource.(*Vpc).Name

	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}

func (r *Vpc) tag(tags map[string]string) error {
	logger.Debug("vpc.Tag")

	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering VPC tag [%s] %s", key, val)
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
