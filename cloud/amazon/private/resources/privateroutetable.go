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

var _ cloud.Resource = &PrivateRouteTable{}

type PrivateRouteTable struct {
	Shared
	ClusterSubnet *cluster.Subnet
	ServerPool    *cluster.ServerPool
}

func (r *PrivateRouteTable) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("privateroutetable.Actual")

	newResource := &PrivateRouteTable{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if r.ClusterSubnet.Identifier != "" {
		output, err := Sdk.Ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			Filters: []*ec2.Filter{
				{
					Name:   S("tag:kubicorn-private-route-table-subnet-pair"),
					Values: []*string{&r.ClusterSubnet.Name},
				},
				{
					Name:   S("vpc-id"),
					Values: []*string{&immutable.ProviderConfig().Network.Identifier},
				},
			},
		})
		if err != nil {
			return nil, nil, err
		}
		lrt := len(output.RouteTables)
		if lrt != 1 {
			return nil, nil, fmt.Errorf("Found [%d] Private Route Tables for ID [%s]", lrt, r.ClusterSubnet.Name)
		}
		rt := output.RouteTables[0]

		newResource.Identifier = r.ClusterSubnet.Name
		for _, tag := range rt.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PrivateRouteTable) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("privateroutetable.Expected")

	newResource := &PrivateRouteTable{
		Shared: Shared{
			Identifier: r.ClusterSubnet.Name,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                     r.Name,
				"KubernetesCluster":                        immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name:  "owned",
				"kubicorn-private-route-table-subnet-pair": r.ClusterSubnet.Name,
			},
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PrivateRouteTable) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("privateroutetable.Apply")

	applyResource := expected.(*PrivateRouteTable)
	isEqual, err := compare.IsEqual(actual.(*PrivateRouteTable), applyResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	rtOutput, err := Sdk.Ec2.CreateRouteTable(&ec2.CreateRouteTableInput{
		VpcId: &immutable.ProviderConfig().Network.Identifier,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Private Route Table: %v", err)
	}
	logger.Success("Created Private Route Table [%s]", *rtOutput.RouteTable.RouteTableId)

	subnetID := ""
	subnetZone := ""
	for _, sp := range immutable.ServerPools() {
		if sp.Name == r.Name {
			for _, sn := range sp.Subnets {
				if sn.Name == r.Name {
					subnetID = sn.Identifier
					subnetZone = sn.Zone
				}
			}
		}
	}
	for _, sn := range immutable.ProviderConfig().Network.PrivateSubnets {
		if sn.Name == r.Name {
			subnetID = sn.Identifier
			subnetZone = sn.Zone
		}
	}
	if subnetID == "" {
		return nil, nil, fmt.Errorf("Unable to find Subnet ID")
	}

	publicSubnetID := ""
	for _, sn := range immutable.ProviderConfig().Network.PublicSubnets {
		if sn.Zone == subnetZone {
			publicSubnetID = sn.Identifier
		}
	}
	if publicSubnetID == "" {
		return nil, nil, fmt.Errorf("Unable to find Public Subnet ID for Zone [%s]", subnetZone)
	}

	output, err := Sdk.Ec2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
		Filter: []*ec2.Filter{
			{
				Name:   S("subnet-id"),
				Values: []*string{&publicSubnetID},
			},
			{
				Name:   S("tag-key"),
				Values: []*string{S("kubicorn-nat-gateway")},
			},
			{
				Name:   S("vpc-id"),
				Values: []*string{&immutable.ProviderConfig().Network.Identifier},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	lng := len(output.NatGateways)
	if lng != 1 {
		return nil, nil, fmt.Errorf("Found [%d] NAT Gateways for Zone [%s]", lng, subnetZone)
	}

	_, err = Sdk.Ec2.CreateRoute(&ec2.CreateRouteInput{
		DestinationCidrBlock: S("0.0.0.0/0"),
		NatGatewayId:         output.NatGateways[0].NatGatewayId,
		RouteTableId:         rtOutput.RouteTable.RouteTableId,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Route: %v", err)
	}

	asInput := &ec2.AssociateRouteTableInput{
		SubnetId:     &subnetID,
		RouteTableId: rtOutput.RouteTable.RouteTableId,
	}
	_, err = Sdk.Ec2.AssociateRouteTable(asInput)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to associate Private Route Table with Subnet: %v", err)
	}
	logger.Success("Associated Route Table [%s] with Subnet [%s]", *rtOutput.RouteTable.RouteTableId, subnetID)

	newResource := &PrivateRouteTable{
		Shared: Shared{
			Identifier: *rtOutput.RouteTable.RouteTableId,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
	}
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new Private Route Table: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PrivateRouteTable) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("privateroutetable.Delete")

	deleteResource := actual.(*PrivateRouteTable)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete Private Route Table resource without ID [%s]", deleteResource.Name)
	}

	output, err := Sdk.Ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   S("tag:kubicorn-private-route-table-subnet-pair"),
				Values: []*string{&r.ClusterSubnet.Name},
			},
			{
				Name:   S("vpc-id"),
				Values: []*string{&immutable.ProviderConfig().Network.Identifier},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	lrt := len(output.RouteTables)
	if lrt != 1 {
		return nil, nil, fmt.Errorf("Found [%d] Private Route Tables for ID [%s]", lrt, deleteResource.Identifier)
	}
	rt := output.RouteTables[0]

	_, err = Sdk.Ec2.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{
		AssociationId: rt.Associations[0].RouteTableAssociationId,
	})
	if err != nil {
		return nil, nil, err
	}

	_, err = Sdk.Ec2.DeleteRouteTable(&ec2.DeleteRouteTableInput{
		RouteTableId: rt.RouteTableId,
	})
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted Private Route Table [%s]", *rt.RouteTableId)

	newResource := &PrivateRouteTable{
		Shared: Shared{
			Name: deleteResource.Name,
			Tags: deleteResource.Tags,
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PrivateRouteTable) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("privateroutetable.Render")

	return inaccurateCluster
}

func (r *PrivateRouteTable) tag(tags map[string]string) error {
	logger.Debug("privateroutetable.Tag")

	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering Private Route Table tag [%s] %s", key, val)
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
