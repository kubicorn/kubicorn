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
					Values: []*string{S(r.ClusterSubnet.Name)},
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

		for _, tag := range rt.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
		newResource.Identifier = r.ClusterSubnet.Name
		newResource.Name = r.ClusterSubnet.Name
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *RouteTable) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("routetable.Expected")
	newResource := &RouteTable{
		Shared: Shared{
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
				"kubicorn-route-table-subnet-pair":        r.ClusterSubnet.Name,
			},
			Identifier: r.ClusterSubnet.Name,
			Name:       r.Name,
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
		return nil, nil, err
	}
	logger.Success("Created Route Table [%s]", *rtOutput.RouteTable.RouteTableId)

	igInput := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []*string{S(immutable.ProviderConfig().Network.InternetGW.Identifier)},
	}
	output, err := Sdk.Ec2.DescribeInternetGateways(igInput)
	if err != nil {
		return nil, nil, err
	}
	lig := len(output.InternetGateways)
	if lig != 1 {
		return nil, nil, fmt.Errorf("Found [%d] Internet Gateways for ID [%s]", lig, immutable.ProviderConfig().Network.InternetGW.Identifier)
	}

	rInput := &ec2.CreateRouteInput{
		DestinationCidrBlock: S("0.0.0.0/0"),
		GatewayId:            output.InternetGateways[0].InternetGatewayId,
		RouteTableId:         rtOutput.RouteTable.RouteTableId,
	}
	_, err = Sdk.Ec2.CreateRoute(rInput)
	if err != nil {
		return nil, nil, err
	}

	subnetID := ""
	for _, sp := range immutable.ServerPools() {
		if sp.Name == r.Name {
			for _, sn := range sp.Subnets {
				if sn.Name == r.Name {
					subnetID = sn.Identifier
				}
			}
		}
	}
	if subnetID == "" {
		return nil, nil, fmt.Errorf("Unable to find Subnet ID")
	}

	asInput := &ec2.AssociateRouteTableInput{
		SubnetId:     &subnetID,
		RouteTableId: rtOutput.RouteTable.RouteTableId,
	}
	_, err = Sdk.Ec2.AssociateRouteTable(asInput)
	if err != nil {
		return nil, nil, err
	}

	logger.Success("Associated Route Table [%s] with Subnet [%s]", *rtOutput.RouteTable.RouteTableId, subnetID)

	newResource := &RouteTable{
		Shared: Shared{
			Tags: make(map[string]string),
		},
	}
	newResource.Identifier = *rtOutput.RouteTable.RouteTableId
	newResource.Name = applyResource.Name

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
				Values: []*string{S(r.ClusterSubnet.Name)},
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

	newResource := &PublicRouteTable{}
	newResource.Name = deleteResource.Name
	newResource.Tags = deleteResource.Tags

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
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

func (r *RouteTable) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("routetable.Render")
	return inaccurateCluster
}
