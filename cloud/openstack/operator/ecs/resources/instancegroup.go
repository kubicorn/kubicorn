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
	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	nets "github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/cloud/openstack/operator/generic/resources"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/script"
)

var _ cloud.Resource = &InstanceGroup{}

//TODO: OperatorPublicNet must be parameterized
const (
	InstancePollingAttempts = 40
	InstancePollingInterval = 5 * time.Second
	OperatorPublicNet       = "PublicNetwork"
	IPv4                    = 4
	IPv6                    = 6
	Fixed                   = "fixed"
	Floating                = "floating"
)

type InstanceGroup struct {
	resources.Shared
	Count            int
	Flavor           string
	Image            string
	BootstrapScripts []string
	ServerPool       *cluster.ServerPool
}

func (r *InstanceGroup) Actual(immutable *cluster.Cluster) (actual *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("instanceGroup.Actual")
	newResource := &InstanceGroup{
		Shared: resources.Shared{
			Name: r.Name,
		},
	}

	// Find instances by name prefix
	res := servers.List(resources.Sdk.Compute, servers.ListOpts{
		Name: r.Name,
	})
	if res.Err != nil {
		return nil, nil, res.Err
	}
	err = res.EachPage(func(page pagination.Page) (bool, error) {
		list, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}
		newResource.Count += len(list)
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}
	newResource.BootstrapScripts = r.ServerPool.BootstrapScripts
	newResource.Flavor = r.ServerPool.Size
	newResource.Image = r.ServerPool.Image

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InstanceGroup) Expected(immutable *cluster.Cluster) (expected *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("instanceGroup.Expected")
	newResource := &InstanceGroup{
		Shared: resources.Shared{
			Name: r.Name,
		},
		BootstrapScripts: r.ServerPool.BootstrapScripts,
		Flavor:           r.ServerPool.Size,
		Image:            r.ServerPool.Image,
		Count:            r.ServerPool.MaxCount,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InstanceGroup) Apply(actual cloud.Resource, expected cloud.Resource, immutable *cluster.Cluster) (updatedCluster *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("instanceGroup.Apply")
	instanceGroup := expected.(*InstanceGroup)
	isEqual, err := compare.IsEqual(actual.(*InstanceGroup), expected.(*InstanceGroup))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, instanceGroup, nil
	}

	var (
		secgroups       []string
		networks        []servers.Network
		masterPublicIP  string
		masterPrivateIP string
	)

	// Wait for master IP
	if r.ServerPool.Type == cluster.ServerPoolTypeNode {
		for i := 0; i < InstancePollingAttempts; i++ {
			masterPublicIP, masterPrivateIP, err = getMasterIPs(immutable)
			if err != nil {
				return nil, nil, err
			}
			if masterPrivateIP == "" || masterPublicIP == "" {
				logger.Info("Waiting for master(s) to come up..")
				time.Sleep(InstancePollingInterval)
				continue
			}
			providerConfig := immutable.ProviderConfig()
			providerConfig.Values.ItemMap["INJECTEDMASTER"] = fmt.Sprintf("%s:%s", masterPrivateIP, immutable.ProviderConfig().KubernetesAPI.Port)
			immutable.SetProviderConfig(providerConfig)
			break
		}
		if _, ok := immutable.ProviderConfig().Values.ItemMap["INJECTEDMASTER"]; !ok {
			return nil, nil, fmt.Errorf("Unable to find Master IP")
		}
	}

	providerConfig := immutable.ProviderConfig()
	providerConfig.Values.ItemMap["INJECTEDPORT"] = immutable.ProviderConfig().KubernetesAPI.Port
	immutable.SetProviderConfig(providerConfig)

	// Build scripts to inject in instance user-data
	userData, err := script.BuildBootstrapScript(r.ServerPool.BootstrapScripts, immutable)
	if err != nil {
		return nil, nil, err
	}

	// Security groups the instances will be part of
	for _, fw := range r.ServerPool.Firewalls {
		secgroups = append(secgroups, fw.Name)
	}

	// Networks instances will be attached to
	networks = append(networks, servers.Network{
		UUID: immutable.ProviderConfig().Network.Identifier,
	})

	// Create instances for this group
	for j := actual.(*InstanceGroup).Count; j < instanceGroup.Count; j++ {
		hostname := fmt.Sprintf("%s-%d", instanceGroup.Name, j+1)
		res := servers.Create(resources.Sdk.Compute, keypairs.CreateOptsExt{
			CreateOptsBuilder: servers.CreateOpts{
				Name:           hostname,
				FlavorName:     instanceGroup.Flavor,
				ImageName:      instanceGroup.Image,
				UserData:       userData,
				SecurityGroups: secgroups,
				Networks:       networks,
				ServiceClient:  resources.Sdk.Compute,
			},
			KeyName: immutable.ProviderConfig().SSH.Name,
		})
		instance, err := res.Extract()
		if err != nil {
			return nil, nil, res.Err
		}
		//Attach Floating IP to the master
		if r.ServerPool.Type == cluster.ServerPoolTypeMaster {
			router := immutable.ProviderConfig().Network.InternetGW
			if router.Identifier != "" {
				masterPublicIP, err = addFloatingIP(instance.ID)
				if err != nil {
					logger.Debug("Failed adding floating ip to instance [%s] with error: %s", instance.ID, err.Error)
				}
				logger.Debug("Cluster endpoint is %s", masterPublicIP)
				providerConfig := immutable.ProviderConfig()
				providerConfig.KubernetesAPI.Endpoint = masterPublicIP
				immutable.SetProviderConfig(providerConfig)
			}
		}
		logger.Debug("Created instance [%s]", instance.ID)
	}

	// If it's not a group of master nodes, wait for them to be up
	if r.ServerPool.Type == cluster.ServerPoolTypeNode {
		var created bool
		for i := 0; i < InstancePollingAttempts; i++ {
			reached, _, err := reachedStatus(instanceGroup.Name, "ACTIVE")
			if err != nil {
				return nil, nil, err
			}
			if reached {
				created = true
				break
			}
			logger.Info("Waiting for node(s) to come up..")
			time.Sleep(InstancePollingInterval)
		}
		if !created {
			return nil, nil, fmt.Errorf("Failed creating instances")
		}
	}

	logger.Success("Created InstanceGroup [%s]", instanceGroup.Name)

	newResource := &InstanceGroup{
		Shared: resources.Shared{
			Name: instanceGroup.Name,
		},
		Image:            instanceGroup.Image,
		Flavor:           instanceGroup.Flavor,
		Count:            instanceGroup.Count,
		BootstrapScripts: instanceGroup.BootstrapScripts,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InstanceGroup) Delete(actual cloud.Resource, immutable *cluster.Cluster) (updatedCluster *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("instanceGroup.Delete")
	instanceGroup := actual.(*InstanceGroup)
	res := servers.List(resources.Sdk.Compute, servers.ListOpts{
		Name: instanceGroup.Name,
	})
	if res.Err != nil {
		return nil, nil, res.Err
	}
	// Remove instances of the group, using the name prefix
	err = res.EachPage(func(page pagination.Page) (bool, error) {
		list, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}

		for _, instance := range list {
			if err := servers.Delete(resources.Sdk.Compute, instance.ID).ExtractErr(); err != nil {
				return false, err
			}
			logger.Debug("Deleting instance [%s]", instance.ID)
		}
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Wait for actual deletion
	var deleted bool
	for i := 0; i < InstancePollingAttempts; i++ {
		count := 0
		res := servers.List(resources.Sdk.Compute, servers.ListOpts{
			Name: instanceGroup.Name,
		})
		if res.Err != nil {
			return nil, nil, err
		}
		err = res.EachPage(func(page pagination.Page) (bool, error) {
			list, err := servers.ExtractServers(page)
			if err != nil {
				return false, err
			}
			count += len(list)
			return false, nil
		})
		if count == 0 {
			deleted = true
			break
		}
		logger.Debug("Waiting for instances to be deleted..")
		time.Sleep(InstancePollingInterval)
	}
	if !deleted {
		return nil, nil, fmt.Errorf("Failed deleting instances")
	}

	logger.Success("Deleted InstanceGroup [%s]", instanceGroup.Name)

	immutable.ProviderConfig().KubernetesAPI.Endpoint = ""

	newResource := &InstanceGroup{
		Shared: resources.Shared{
			Name: instanceGroup.Name,
		},
		BootstrapScripts: instanceGroup.BootstrapScripts,
		Flavor:           instanceGroup.Flavor,
		Image:            instanceGroup.Image,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *InstanceGroup) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("instanceGroup.Render")
	instanceGroup := newResource.(*InstanceGroup)
	newCluster := inaccurateCluster
	found := false
	machineProviderConfigs := newCluster.MachineProviderConfigs()
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		pool := machineProviderConfig.ServerPool
		if pool.Name == instanceGroup.Name {
			pool.Image = instanceGroup.Image
			pool.Size = instanceGroup.Flavor
			pool.BootstrapScripts = instanceGroup.BootstrapScripts
			found = true
			machineProviderConfig.ServerPool = pool
			machineProviderConfigs[i] = machineProviderConfig
			newCluster.SetMachineProviderConfigs(machineProviderConfigs)
		}
	}
	if !found {
		providerConfig := []*cluster.MachineProviderConfig{
			{
				ServerPool: &cluster.ServerPool{
					Name:             instanceGroup.Name,
					BootstrapScripts: instanceGroup.BootstrapScripts,
					Image:            instanceGroup.Image,
					Size:             instanceGroup.Flavor,
				},
			},
		}
		newCluster.NewMachineSetsFromProviderConfigs(providerConfig)
	}
	return newCluster
}

func addFloatingIP(serverID string) (string, error) {
	extNetID, err := nets.IDFromName(resources.Sdk.Network, OperatorPublicNet)
	if err != nil {
		return "", fmt.Errorf("Unable to find external network ID")
	}
	fip, err := floatingips.Create(resources.Sdk.Network, floatingips.CreateOpts{
		FloatingNetworkID: extNetID,
	}).Extract()

	listOpts := ports.ListOpts{
		DeviceID: serverID,
	}
	allPages, err := ports.List(resources.Sdk.Network, listOpts).AllPages()
	if err != nil {
		return "", err
	}
	allPorts, err := ports.ExtractPorts(allPages)
	if err != nil {
		return "", err
	}
	// one single port is expected
	if len(allPorts) != 1 {
		return "", fmt.Errorf("Unable to find the expected number of ports")
	}
	portID := allPorts[0].ID
	fip, err = floatingips.Update(resources.Sdk.Network, fip.ID, floatingips.UpdateOpts{
		PortID: &portID,
	}).Extract()
	if err != nil {
		return "", err
	}
	return fip.FloatingIP, err
}

func getMasterIPs(immutable *cluster.Cluster) (string, string, error) {
	var masterName string
	for _, pool := range immutable.ServerPools() {
		if pool.Type == cluster.ServerPoolTypeMaster {
			masterName = pool.Name
			break
		}
	}
	reached, instances, err := reachedStatus(masterName, "ACTIVE")
	if err != nil {
		return "", "", err
	}
	if !reached {
		return "", "", err
	}

	publicIP := getNetworkIP(instances[0], immutable.ProviderConfig().Network.Name, IPv4, Floating)
	privateIP := getNetworkIP(instances[0], immutable.ProviderConfig().Network.Name, IPv4, Fixed)
	return publicIP, privateIP, err
}

func reachedStatus(hostname, status string) (reached bool, instances []servers.Server, err error) {
	var total int
	res := servers.List(resources.Sdk.Compute, servers.ListOpts{
		Name: hostname,
	})
	if res.Err != nil {
		err = res.Err
		return
	}
	err = res.EachPage(func(page pagination.Page) (bool, error) {
		list, lerr := servers.ExtractServers(page)
		if lerr != nil {
			return false, nil
		}
		total += len(list)
		for _, server := range list {
			if server.Status == status {
				instances = append(instances, server)
			}
		}
		return true, nil
	})
	if total == len(instances) {
		reached = true
		return
	}
	return
}

func getNetworkIP(srv servers.Server, network string, version int, iptype string) (IP string) {
	for netName, v := range srv.Addresses {
		if netName != network {
			continue
		}
		for _, v := range v.([]interface{}) {
			v := v.(map[string]interface{})
			if v["OS-EXT-IPS:type"] == iptype {
				if v["version"].(float64) == float64(version) {
					return v["addr"].(string)
				}
			}
		}
	}
	return ""
}
