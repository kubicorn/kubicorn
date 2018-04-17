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

var _ cloud.Resource = &RouteTable{}

type RouteTable struct {
	Shared
	ClusterSubnet *cluster.Subnet
	ServerPool    *cluster.ServerPool
}

func (r *RouteTable) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("routetable.Actual")
	newResource := &RouteTable{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if r.ClusterSubnet.Identifier != "" {
		input := &ec2.DescribeRouteTablesInput{
			Filters: []*ec2.Filter{
				{
					Name:   S("tag:kubicorn-route-table-subnet-pair"),
					Values: []*string{&r.ClusterSubnet.Name},
				},
				{
					Name:   S("vpc-id"),
					Values: []*string{&immutable.ProviderConfig().Network.Identifier},
				},
			},
		}
		output, err := Sdk.Ec2.DescribeRouteTables(input)
		if err != nil {
			return nil, nil, err
		}
		lrt := len(output.RouteTables)
		if lrt != 1 {
			return nil, nil, fmt.Errorf("Found [%d] Route Tables for ID [%s]", lrt, r.ClusterSubnet.Name)
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

func (r *RouteTable) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("routetable.Expected")
	newResource := &RouteTable{
		Shared: Shared{
			Identifier: r.ClusterSubnet.Name,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
				"kubicorn-route-table-subnet-pair":        r.ClusterSubnet.Name,
			},
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *RouteTable) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("routetable.Apply")
	applyResource := expected.(*RouteTable)
	isEqual, err := compare.IsEqual(actual.(*RouteTable), applyResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	rtInput := &ec2.CreateRouteTableInput{
		VpcId: &immutable.ProviderConfig().Network.Identifier,
	}
	rtOutput, err := Sdk.Ec2.CreateRouteTable(rtInput)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Route Table: %v", err)
	}
	logger.Success("Created Route Table [%s]", *rtOutput.RouteTable.RouteTableId)

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
		return nil, nil, fmt.Errorf("Unable to find Public Subnet ID for Zone [%s}", subnetZone)
	}

	ngInput := &ec2.DescribeNatGatewaysInput{
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
	}
	output, err := Sdk.Ec2.DescribeNatGateways(ngInput)
	if err != nil {
		return nil, nil, err
	}
	lng := len(output.NatGateways)
	if lng != 1 {
		return nil, nil, fmt.Errorf("Found [%d] NAT Gateways for Zone [%s]", lng, subnetZone)
	}

	rInput := &ec2.CreateRouteInput{
		DestinationCidrBlock: S("0.0.0.0/0"),
		NatGatewayId:         output.NatGateways[0].NatGatewayId,
		RouteTableId:         rtOutput.RouteTable.RouteTableId,
	}
	_, err = Sdk.Ec2.CreateRoute(rInput)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Route: %v", err)
	}

	asInput := &ec2.AssociateRouteTableInput{
		SubnetId:     &subnetID,
		RouteTableId: rtOutput.RouteTable.RouteTableId,
	}
	_, err = Sdk.Ec2.AssociateRouteTable(asInput)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to associate Route Table with Subnet: %v", err)
	}
	logger.Success("Associated Route Table [%s] with Subnet [%s]", *rtOutput.RouteTable.RouteTableId, subnetID)

	newResource := &RouteTable{
		Shared: Shared{
			Identifier: *rtOutput.RouteTable.RouteTableId,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
	}
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new Route Table: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *RouteTable) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("routetable.Delete")
	deleteResource := actual.(*RouteTable)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete Route Table resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   S("tag:kubicorn-route-table-subnet-pair"),
				Values: []*string{&r.ClusterSubnet.Name},
			},
			{
				Name:   S("vpc-id"),
				Values: []*string{&immutable.ProviderConfig().Network.Identifier},
			},
		},
	}
	output, err := Sdk.Ec2.DescribeRouteTables(input)
	if err != nil {
		return nil, nil, err
	}
	lrt := len(output.RouteTables)
	if lrt != 1 {
		return nil, nil, fmt.Errorf("Found [%d] Route Tables for ID [%s]", lrt, deleteResource.Identifier)
	}
	rt := output.RouteTables[0]

	daInput := &ec2.DisassociateRouteTableInput{
		AssociationId: rt.Associations[0].RouteTableAssociationId,
	}
	_, err = Sdk.Ec2.DisassociateRouteTable(daInput)
	if err != nil {
		return nil, nil, err
	}

	dInput := &ec2.DeleteRouteTableInput{
		RouteTableId: rt.RouteTableId,
	}
	_, err = Sdk.Ec2.DeleteRouteTable(dInput)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted Route Table [%s]", deleteResource.Identifier)

	newResource := &PublicRouteTable{
		Shared: Shared{
			Name: deleteResource.Name,
			Tags: deleteResource.Tags,
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *RouteTable) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("routetable.Render")

	return inaccurateCluster
}

func (r *RouteTable) tag(tags map[string]string) error {
	logger.Debug("routetable.Tag")
	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering Route Table tag [%s] %s", key, val)
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
