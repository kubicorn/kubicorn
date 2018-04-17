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

var _ cloud.Resource = &PublicRouteTable{}

type PublicRouteTable struct {
	Shared
	ClusterSubnet *cluster.Subnet
}

func (r *PublicRouteTable) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicroutetable.Actual")

	newResource := &PublicRouteTable{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if r.ClusterSubnet.Identifier != "" {
		output, err := Sdk.Ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			Filters: []*ec2.Filter{
				{
					Name:   S("tag:kubicorn-public-route-table-subnet-pair"),
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
			return nil, nil, fmt.Errorf("Found [%d] Public Route Tables for ID [%s]", lrt, r.ClusterSubnet.Name)
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

func (r *PublicRouteTable) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicroutetable.Expected")

	newResource := &PublicRouteTable{
		Shared: Shared{
			Identifier: r.ClusterSubnet.Name,
			Name:       r.Name,
			Tags: map[string]string{
				"Name":                                    r.Name,
				"KubernetesCluster":                       immutable.Name,
				"kubernetes.io/cluster/" + immutable.Name: "owned",
				"kubicorn-public-route-table-subnet-pair": r.ClusterSubnet.Name,
			},
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicRouteTable) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicroutetable.Apply")

	applyResource := expected.(*PublicRouteTable)
	isEqual, err := compare.IsEqual(actual.(*PublicRouteTable), applyResource)
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
		return nil, nil, fmt.Errorf("Unable to create new Public Route Table: %v", err)
	}
	logger.Success("Created Public Route Table [%s]", *rtOutput.RouteTable.RouteTableId)

	subnetID := ""
	for _, sn := range immutable.ProviderConfig().Network.PublicSubnets {
		if sn.Name == r.Name {
			subnetID = sn.Identifier
		}
	}
	if subnetID == "" {
		return nil, nil, fmt.Errorf("Unable to find Public Subnet ID")
	}

	_, err = Sdk.Ec2.CreateRoute(&ec2.CreateRouteInput{
		DestinationCidrBlock: S("0.0.0.0/0"),
		GatewayId:            &immutable.ProviderConfig().Network.InternetGW.Identifier,
		RouteTableId:         rtOutput.RouteTable.RouteTableId,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create new Route: %v", err)
	}

	_, err = Sdk.Ec2.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		SubnetId:     &subnetID,
		RouteTableId: rtOutput.RouteTable.RouteTableId,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to associate Public Route Table with Subnet: %v", err)
	}
	logger.Success("Associated Public Route Table [%s] with Public Subnet [%s]", *rtOutput.RouteTable.RouteTableId, subnetID)

	newResource := &PublicRouteTable{
		Shared: Shared{
			Identifier: *rtOutput.RouteTable.RouteTableId,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
	}
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new Public Route Table: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicRouteTable) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("publicroutetable.Delete")

	deleteResource := actual.(*PublicRouteTable)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete Public Route Table resource without ID [%s]", deleteResource.Name)
	}

	output, err := Sdk.Ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   S("tag:kubicorn-public-route-table-subnet-pair"),
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
		return nil, nil, fmt.Errorf("Found [%d] Public Route Tables for ID [%s]", lrt, deleteResource.Identifier)
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
	logger.Success("Deleted Public Route Table [%s]", *rt.RouteTableId)

	newResource := &PublicRouteTable{
		Shared: Shared{
			Name: deleteResource.Name,
			Tags: deleteResource.Tags,
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *PublicRouteTable) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("publicroutetable.Render")

	return inaccurateCluster
}

func (r *PublicRouteTable) tag(tags map[string]string) error {
	logger.Debug("publicroutetable.Tag")

	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering Public Route Table tag [%s] %s", key, val)
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
