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

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/defaults"
	"github.com/kris-nova/kubicorn/cutil/logger"
)

var _ cloud.Resource = &LoadBalancer{}

type LoadBalancer struct {
	Shared
	ServerPool     *cluster.ServerPool
	Subnet         *cluster.Subnet
	BackendPoolIDs []string
	NatPoolIDs     []string
}

func (r *LoadBalancer) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("loadbalancer.Actual")
	newResource := &LoadBalancer{
		Shared: Shared{
			Tags:       r.Tags,
			Identifier: immutable.Network.Identifier,
		},
	}

	lb, err := Sdk.LoadBalancer.Get(immutable.Name, r.ServerPool.Name, "")
	if err != nil {
		logger.Debug("Error looking up load balancer [%s]: %v", r.ServerPool.Name, err)
	} else {
		newResource.Name = *lb.Name
		for _, b := range *lb.BackendAddressPools {
			newResource.BackendPoolIDs = append(newResource.BackendPoolIDs, *b.ID)
		}
		for _, b := range *lb.InboundNatPools {
			newResource.NatPoolIDs = append(newResource.NatPoolIDs, *b.ID)
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *LoadBalancer) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("loadbalancer.Expected")
	newResource := &LoadBalancer{
		Shared: Shared{
			Name:       r.Name,
			Tags:       r.Tags,
			Identifier: immutable.Network.Identifier,
		},
		NatPoolIDs:     r.Subnet.LoadBalancer.NATIDs,
		BackendPoolIDs: r.Subnet.LoadBalancer.BackendIDs,
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *LoadBalancer) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("loadbalancer.Apply")
	applyResource := expected.(*LoadBalancer)
	isEqual, err := compare.IsEqual(actual.(*LoadBalancer), expected.(*LoadBalancer))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	parameters := network.LoadBalancer{
		Location: &immutable.Location,
	}
	lbch, errch := Sdk.LoadBalancer.CreateOrUpdate(immutable.Name, r.ServerPool.Name, parameters, make(chan struct{}))
	lb := <-lbch
	err = <-errch
	if err != nil {
		return nil, nil, err
	}
	logger.Info("Created or updated load balancer [%s]", *lb.ID)
	var backEndPools []string
	for _, b := range *lb.BackendAddressPools {
		backEndPools = append(backEndPools, *b.ID)
	}
	var inboundNatPools []string
	for _, b := range *lb.BackendAddressPools {
		inboundNatPools = append(inboundNatPools, *b.ID)
	}

	newResource := &LoadBalancer{
		Shared: Shared{
			Name:       *lb.Name,
			Tags:       r.Tags,
			Identifier: immutable.Network.Identifier,
		},
		NatPoolIDs:     inboundNatPools,
		BackendPoolIDs: backEndPools,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}
func (r *LoadBalancer) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("loadbalancer.Delete")
	deleteResource := actual.(*LoadBalancer)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete VPC resource without ID [%s]", deleteResource.Name)
	}

	respch, errch := Sdk.LoadBalancer.Delete(immutable.Name, r.ServerPool.Name, make(chan struct{}))
	<-respch
	err := <-errch
	if err != nil {
		return nil, nil, nil
	}
	logger.Info("Deleted load balancer [%s]", deleteResource.Identifier)
	newResource := &LoadBalancer{
		Shared: Shared{
			Name:       r.ServerPool.Name,
			Tags:       r.Tags,
			Identifier: immutable.Network.Identifier,
		},
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *LoadBalancer) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("loadbalancer.Render")
	newCluster := defaults.NewClusterDefaults(inaccurateCluster)
	for _, serverPool := range newCluster.ServerPools {
		for _, subnet := range serverPool.Subnets {
			if subnet.LoadBalancer.Name == newResource.(*LoadBalancer).Name {
				subnet.LoadBalancer.BackendIDs = newResource.(*LoadBalancer).BackendPoolIDs
				subnet.LoadBalancer.NATIDs = newResource.(*LoadBalancer).NatPoolIDs
			}
		}
	}
	return newCluster
}
