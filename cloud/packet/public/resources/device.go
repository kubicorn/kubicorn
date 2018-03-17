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

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/script"
	"github.com/packethost/packngo"
)

const (
	MasterIPAttempts               = 100
	MasterIPSleepSecondsPerAttempt = 5
)

var _ cloud.Resource = &Device{}

type Device struct {
	Shared
	Location         string
	Type             string
	OS               string
	SSHFingerprint   string
	BootstrapScripts []string
	Count            int
	ServerPool       *cluster.ServerPool
	Tags             []string
	ProjectID        string
}

func (r *Device) String() string {
	return fmt.Sprintf(
		"shared:%v location:%s type:%s os:%s sshfingerprint:%s bootstrap:%s count:%d tags:%s projectid:%s",
		r.Shared,
		r.Location,
		r.Type, r.OS, r.SSHFingerprint, r.BootstrapScripts, r.Count, r.Tags, r.ProjectID)
}

func (r *Device) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("device.Actual")
	newResource := &Device{
		Shared: Shared{
			Name: r.Name,
		},
	}

	var valid []packngo.Device
	logger.Debug("device.Actual finding project ID by name %s", immutable.ProviderConfig().Project.Name)
	project, err := GetProjectByName(immutable.ProviderConfig().Project.Name)
	if err != nil {
		return nil, nil, err
	}
	if project != nil && project.ID != "" {
		tag := r.ServerPool.Type
		logger.Debug("device.Actual getting devices with tag %s", tag)
		valid, err = devicesByTag(project.ID, tag)
		logger.Debug("device.Actual devices found %v", valid)
		if err != nil {
			return nil, nil, err
		}
	}

	if len(valid) > 0 {
		newResource.Count = len(valid)
		device := valid[0]
		newResource.Name = r.Name
		newResource.Type = device.Plan.Slug
		newResource.OS = device.OS.Slug
		newResource.Location = device.Facility.Code
		newResource.BootstrapScripts = r.ServerPool.BootstrapScripts
		newResource.SSHFingerprint = immutable.ProviderConfig().SSH.PublicKeyFingerprint
		newResource.Tags = device.Tags
		newResource.ProjectID = project.ID

	} else {
		newResource.Count = 0
		newResource.Type = r.ServerPool.Size
		newResource.OS = r.ServerPool.Image
		newResource.Location = immutable.ProviderConfig().Location
		newResource.BootstrapScripts = r.ServerPool.BootstrapScripts
		newResource.SSHFingerprint = immutable.ProviderConfig().SSH.PublicKeyFingerprint
		newResource.Name = r.ServerPool.Name
		newResource.Tags = r.Tags
	}

	logger.Debug("device.Actual newResource %v", newResource)

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Device) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("device.Expected")
	newResource := &Device{
		Shared: Shared{
			Name: r.Name,
		},
		Type:             r.ServerPool.Size,
		Location:         immutable.ProviderConfig().Location,
		OS:               r.ServerPool.Image,
		SSHFingerprint:   immutable.ProviderConfig().SSH.PublicKeyFingerprint,
		BootstrapScripts: r.ServerPool.BootstrapScripts,
		Count:            r.ServerPool.MaxCount,
		Tags:             []string{r.ServerPool.Type},
		ProjectID:        r.ProjectID,
	}
	logger.Debug("device.Expected newResource %v", newResource)

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Device) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("device.Apply")
	expectedResource := expected.(*Device)
	actualResource := actual.(*Device)
	// do we know our project ID?
	projectID := actualResource.ProjectID
	logger.Debug("device.Apply project ID from actual [%s]", projectID)
	if projectID == "" {
		logger.Debug("device.Apply retrieving project ID for [%s]", immutable.ProviderConfig().Project.Name)
		project, err := GetProjectByName(immutable.ProviderConfig().Project.Name)
		if err != nil {
			return nil, nil, err
		}
		if project != nil {
			projectID = project.ID
		}
	}
	// take 2
	logger.Debug("device.Apply project ID ready for processing [%s]", projectID)
	if projectID == "" {
		return nil, nil, fmt.Errorf("Cannot work without valid project ID")
	}
	// we can add the projectID to the expected
	expectedResource.ProjectID = projectID
	logger.Debug("device.Apply expected vs actual %v %v", *expectedResource, *actualResource)
	// just copy over the ID if it exists
	isEqual, err := compare.IsEqual(actualResource, expectedResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		logger.Debug("device.Apply already equal")
		return immutable, expectedResource, nil
	}

	// create the device

	// if we are a node, we need to get thekubernetes master IP
	if r.ServerPool.Type == cluster.ServerPoolTypeNode {
		masterIPs, err := getMasterIP(projectID, cluster.ServerPoolTypeMaster)
		if err != nil {
			return nil, nil, err
		}

		found := masterIPs[0] != "" && masterIPs[1] != "" && masterIPs[2] != ""
		logger.Debug("device.Actual master IP addresses %t %s %s %s", found, masterIPs[0], masterIPs[1], masterIPs[2])

		if !found {
			return nil, nil, fmt.Errorf("Unable to find master IP addresses")
		}

		immutable.ProviderConfig().KubernetesAPI.Endpoint = masterIPs[0]
		immutable.ProviderConfig().Values.ItemMap["INJECTEDMASTER"] = fmt.Sprintf("%s:%s", masterIPs[2], immutable.ProviderConfig().KubernetesAPI.Port)
	}

	userData, err := script.BuildBootstrapScript(r.ServerPool.BootstrapScripts, immutable)
	if err != nil {
		return nil, nil, err
	}

	logger.Debug("device.Apply devices actual %d, expected %d", actualResource.Count, expectedResource.Count)
	var device *packngo.Device
	for j := actualResource.Count; j < expectedResource.Count; j++ {
		hostname := fmt.Sprintf("%s-%d", expected.(*Device).Name, j)
		createRequest := &packngo.DeviceCreateRequest{
			HostName:     hostname,
			Facility:     expectedResource.Location,
			Plan:         expectedResource.Type,
			OS:           expectedResource.OS,
			UserData:     string(userData),
			Tags:         expectedResource.Tags,
			ProjectID:    projectID,
			BillingCycle: "hourly",
		}
		logger.Debug("creating server %s: %v", hostname, createRequest)
		device, _, err = Sdk.Client.Devices.Create(createRequest)
		if err != nil {
			return nil, nil, err
		}
		logger.Success("Created Device [%s]", device.Hostname)
	}

	// if these are masters, we are not done until we have the master IP
	if r.ServerPool.Type == cluster.ServerPoolTypeMaster {
		masterIPs, err := getMasterIP(projectID, cluster.ServerPoolTypeMaster)
		if err != nil {
			return nil, nil, err
		}

		found := masterIPs[0] != "" && masterIPs[1] != "" && masterIPs[2] != ""
		logger.Debug("device.Apply master IP addresses %t %s %s %s", found, masterIPs[0], masterIPs[1], masterIPs[2])

		if !found {
			return nil, nil, fmt.Errorf("Unable to find master IP addresses")
		}

		immutable.ProviderConfig().KubernetesAPI.Endpoint = masterIPs[0]

	}

	newResource := &Device{
		Shared: Shared{
			Name: r.ServerPool.Name,
		},
		BootstrapScripts: expected.(*Device).BootstrapScripts,
	}
	if expectedResource.Count != 0 && device != nil {
		newResource.OS = device.OS.Slug
		newResource.Type = device.Plan.Slug
		newResource.Location = device.Facility.Code
		newResource.Count = r.Count
	}

	logger.Debug("device.Apply newResource %v", newResource)

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Device) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("device.Delete")
	deleteResource := actual.(*Device)
	logger.Debug("device.Delete deleteResource %v", deleteResource)
	if deleteResource.Name == "" {
		return nil, nil, fmt.Errorf("Unable to delete device resource without Name [%s]", deleteResource.Name)
	}

	valid, err := devicesByTag(r.ProjectID, r.ServerPool.Type)
	logger.Debug("device.Delete devices to delete %v", valid)
	if err != nil {
		return nil, nil, err
	}

	for _, device := range valid {

		_, err := Sdk.Client.Devices.Delete(device.ID)
		if err != nil {
			return nil, nil, err
		}
		logger.Success("Deleted Device [%s]", device.Hostname)
	}

	// Kubernetes API
	// todo (@kris-nova) this is obviously not immutable
	immutable.ProviderConfig().KubernetesAPI.Endpoint = ""

	newResource := &Device{
		OS:               deleteResource.OS,
		Type:             deleteResource.Type,
		Location:         deleteResource.Location,
		BootstrapScripts: deleteResource.BootstrapScripts,
		Count:            0,
	}
	newResource.Name = deleteResource.Name

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Device) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("device.Render")

	newCluster := inaccurateCluster
	serverPool := &cluster.ServerPool{}
	serverPool.Type = r.ServerPool.Type
	serverPool.Image = newResource.(*Device).OS
	serverPool.Size = newResource.(*Device).Type
	serverPool.Name = newResource.(*Device).Name
	serverPool.BootstrapScripts = newResource.(*Device).BootstrapScripts
	found := false
	machineProviderConfigs := newCluster.MachineProviderConfigs()
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		if machineProviderConfig.Name == newResource.(*Device).Name {
			machineProviderConfig.ServerPool.Image = newResource.(*Device).OS
			machineProviderConfig.ServerPool.Size = newResource.(*Device).Type
			machineProviderConfig.ServerPool.BootstrapScripts = newResource.(*Device).BootstrapScripts
			found = true
			machineProviderConfigs[i] = machineProviderConfig
			newCluster.SetMachineProviderConfigs(machineProviderConfigs)
		}
	}
	if !found {
		providerConfig := []*cluster.MachineProviderConfig{
			{
				ServerPool: serverPool,
			},
		}
		newCluster.NewMachineSetsFromProviderConfigs(providerConfig)
	}
	providerConfig := newCluster.ProviderConfig()
	providerConfig.Location = newResource.(*Device).Location
	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}

func devicesByTag(project, tag string) ([]packngo.Device, error) {
	devices, response, err := Sdk.Client.Devices.List(project)
	if err != nil && response.StatusCode != 404 {
		return nil, err
	}
	valid := make([]packngo.Device, 0, len(devices))
	// now find all of the devices that have the tags for our type
	for _, device := range devices {
		for _, onetag := range device.Tags {
			if onetag == tag {
				valid = append(valid, device)
			}
		}
	}
	return valid, nil
}

func getMasterIP(project, tag string) ([]string, error) {
	// get the masteriP
	ret := make([]string, 3, 3)
	logger.Debug("device.getMasterIP attempting to get master public IP")
	for i := 0; i < MasterIPAttempts; i++ {
		logger.Debug("device.getMasterIP attempt %d to get master IP address", i)
		devices, err := devicesByTag(project, tag)
		if err != nil {
			logger.Debug("device.getMasterIP error retrieving devices: %v", err)
			return ret, err
		}
		// we have master devices
		if len(devices) > 0 {
			ips := devices[0].Network
			if len(ips) > 0 {
				for _, ip := range ips {
					if ip.Address != "" {
						switch {
						case ip.AddressFamily == 4 && ip.Public:
							ret[0] = ip.Address
						case ip.AddressFamily == 6 && ip.Public:
							ret[1] = ip.Address
						case ip.AddressFamily == 4 && !ip.Public:
							ret[2] = ip.Address
						}
					}
				}
			}
		}
		if ret[0] != "" && ret[1] != "" && ret[2] != "" {
			break
		}
		time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
	}
	return ret, nil
}
