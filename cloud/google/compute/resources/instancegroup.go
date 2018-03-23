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

	"strings"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/kubicorn/kubicorn/pkg/script"
	"google.golang.org/api/compute/v1"
)

var _ cloud.Resource = &InstanceGroup{}

// InstanceGroup is a representation of the server to be created on the cloud provider.
type InstanceGroup struct {
	Shared
	Location         string
	Size             string
	Image            string
	Count            int
	SSHFingerprint   string
	BootstrapScripts []string
	ServerPool       *cluster.ServerPool
}

const (
	// MasterIPAttempts specifies how many times are allowed to be taken to get the master node IP.
	MasterIPAttempts = 40
	// MasterIPSleepSecondsPerAttempt specifies how much time should pass after a failed attempt to get the master IP.
	MasterIPSleepSecondsPerAttempt = 3
	// DeleteAttempts specifies the amount of retries are allowed when trying to delete instance templates.
	DeleteAttempts = 150
	// RetrySleepSeconds specifies the time to sleep after a failed attempt to delete instance templates.
	DeleteSleepSeconds = 5
)

// Actual is used to build a cluster based on instances on the cloud provider.
func (r *InstanceGroup) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instanceGroup.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached instance [actual]")
		return immutable, r.CachedActual, nil
	}
	newResource := &InstanceGroup{
		Shared: Shared{
			Name:    r.Name,
			CloudID: r.ServerPool.Identifier,
		},
	}

	project, err := Sdk.Service.Projects.Get(immutable.ProviderConfig().CloudId).Do()
	if err != nil && project != nil {
		instances, err := Sdk.Service.Instances.List(immutable.ProviderConfig().CloudId, immutable.ProviderConfig().Location).Do()
		if err != nil {
			return nil, nil, err
		}

		count := len(instances.Items)
		if count > 0 {
			newResource.Count = count

			instance := instances.Items[0]
			newResource.Name = instance.Name
			newResource.CloudID = string(instance.Id)
			newResource.Size = instance.Kind
			newResource.Image = r.Image
			newResource.Location = instance.Zone
		}
	}

	newResource.BootstrapScripts = r.ServerPool.BootstrapScripts
	newResource.SSHFingerprint = immutable.ProviderConfig().SSH.PublicKeyFingerprint
	newResource.Name = r.Name
	r.CachedActual = newResource
	return immutable, newResource, nil
}

// Expected is used to build a cluster expected to be on the cloud provider.
func (r *InstanceGroup) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instanceGroup.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using instance subnet [expected]")
		return immutable, r.CachedExpected, nil
	}
	expected := &InstanceGroup{
		Shared: Shared{
			Name:    r.Name,
			CloudID: r.ServerPool.Identifier,
		},
		Size:             r.ServerPool.Size,
		Location:         immutable.ProviderConfig().Location,
		Image:            r.ServerPool.Image,
		Count:            r.ServerPool.MaxCount,
		SSHFingerprint:   immutable.ProviderConfig().SSH.PublicKeyFingerprint,
		BootstrapScripts: r.ServerPool.BootstrapScripts,
	}
	r.CachedExpected = expected
	return immutable, expected, nil
}

// Apply is used to create the expected resources on the cloud provider.
func (r *InstanceGroup) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instanceGroup.Apply")
	applyResource := expected.(*InstanceGroup)
	isEqual, err := compare.IsEqual(actual.(*InstanceGroup), expected.(*InstanceGroup))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	masterIPPrivate := ""
	masterIPPublic := ""
	if r.ServerPool.Type == cluster.ServerPoolTypeNode {
		found := false
		for i := 0; i < MasterIPAttempts; i++ {
			masterTag := ""
			machineConfigs := immutable.MachineProviderConfigs()
			for _, machineConfig := range machineConfigs {
				serverPool := machineConfig.ServerPool
				if serverPool.Type == cluster.ServerPoolTypeMaster {
					masterTag = serverPool.Name
				}
			}
			if masterTag == "" {
				return nil, nil, fmt.Errorf("Unable to find master tag")
			}

			instanceGroupManager, err := Sdk.Service.InstanceGroupManagers.ListManagedInstances(immutable.ProviderConfig().CloudId, expected.(*InstanceGroup).Location, strings.ToLower(masterTag)).Do()
			if err != nil {
				return nil, nil, err
			}

			if err != nil || len(instanceGroupManager.ManagedInstances) == 0 {
				logger.Debug("Hanging for master IP.. (%v)", err)
				time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
				continue
			}

			parts := strings.Split(instanceGroupManager.ManagedInstances[0].Instance, "/")
			instance, err := Sdk.Service.Instances.Get(immutable.ProviderConfig().CloudId, expected.(*InstanceGroup).Location, parts[len(parts)-1]).Do()
			if err != nil {
				logger.Debug("Hanging for master IP.. (%v)", err)
				time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
				continue
			}

			for _, networkInterface := range instance.NetworkInterfaces {
				if networkInterface.Name == "nic0" {
					masterIPPrivate = networkInterface.NetworkIP
					for _, accessConfigs := range networkInterface.AccessConfigs {
						masterIPPublic = accessConfigs.NatIP
					}
				}
			}

			if masterIPPublic == "" {
				logger.Debug("Hanging for master IP..")
				time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
				continue
			}

			found = true
			providerConfig := immutable.ProviderConfig()
			providerConfig.Values.ItemMap["INJECTEDMASTER"] = fmt.Sprintf("%s:%s", masterIPPrivate, immutable.ProviderConfig().KubernetesAPI.Port)
			immutable.SetProviderConfig(providerConfig)
			break
		}
		if !found {
			return nil, nil, fmt.Errorf("Unable to find Master IP after defined wait")
		}
	}

	providerConfig := immutable.ProviderConfig()
	providerConfig.Values.ItemMap["INJECTEDPORT"] = immutable.ProviderConfig().KubernetesAPI.Port
	immutable.SetProviderConfig(providerConfig)

	scripts, err := script.BuildBootstrapScript(r.ServerPool.BootstrapScripts, immutable)
	if err != nil {
		return nil, nil, err
	}

	finalScripts := string(scripts)
	if err != nil {
		return nil, nil, err
	}

	tags := []string{}
	if r.ServerPool.Type == cluster.ServerPoolTypeMaster {
		if immutable.ProviderConfig().KubernetesAPI.Port == "443" {
			tags = append(tags, "https-server")
		}

		if immutable.ProviderConfig().KubernetesAPI.Port == "80" {
			tags = append(tags, "http-server")
		}

		tags = append(tags, "kubicorn-master")
	}

	if r.ServerPool.Type == cluster.ServerPoolTypeNode {
		tags = append(tags, "kubicorn-node")
	}

	prefix := "https://www.googleapis.com/compute/v1/projects/" + immutable.ProviderConfig().CloudId
	imageURL := "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/" + expected.(*InstanceGroup).Image

	templateInstance, err := Sdk.Service.InstanceTemplates.Get(immutable.ProviderConfig().CloudId, strings.ToLower(expected.(*InstanceGroup).Name)).Do()
	if err != nil {
		sshPublicKeyValue := fmt.Sprintf("%s:%s", immutable.ProviderConfig().SSH.User, string(immutable.ProviderConfig().SSH.PublicKeyData))

		templateInstance = &compute.InstanceTemplate{
			Name: strings.ToLower(expected.(*InstanceGroup).Name),
			Properties: &compute.InstanceProperties{
				MachineType: expected.(*InstanceGroup).Size,
				Disks: []*compute.AttachedDisk{
					{
						AutoDelete: true,
						Boot:       true,
						Type:       "PERSISTENT",
						InitializeParams: &compute.AttachedDiskInitializeParams{
							SourceImage: imageURL,
						},
					},
				},
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						AccessConfigs: []*compute.AccessConfig{
							{
								Type: "ONE_TO_ONE_NAT",
								Name: "External NAT",
							},
						},
						Network: prefix + "/global/networks/default",
					},
				},
				ServiceAccounts: []*compute.ServiceAccount{
					{
						Email: "default",
						Scopes: []string{
							compute.DevstorageFullControlScope,
							compute.ComputeScope,
						},
					},
				},
				Metadata: &compute.Metadata{
					Kind: "compute#metadata",
					Items: []*compute.MetadataItems{
						{
							Key:   "ssh-keys",
							Value: &sshPublicKeyValue,
						},
						{
							Key:   "startup-script",
							Value: &finalScripts,
						},
					},
				},
				Tags: &compute.Tags{
					Items: tags,
				},
			},
		}

		_, err = Sdk.Service.InstanceTemplates.Insert(immutable.ProviderConfig().CloudId, templateInstance).Do()
		if err != nil {
			return nil, nil, err
		}
	}

	_, err = Sdk.Service.InstanceGroupManagers.Get(immutable.ProviderConfig().CloudId, expected.(*InstanceGroup).Location, strings.ToLower(expected.(*InstanceGroup).Name)).Do()
	if err != nil {
		instanceGroupManager := &compute.InstanceGroupManager{
			Name:             templateInstance.Name,
			BaseInstanceName: templateInstance.Name,
			InstanceTemplate: prefix + "/global/instanceTemplates/" + templateInstance.Name,
			TargetSize:       int64(expected.(*InstanceGroup).Count),
		}

		for i := 0; i < MasterIPAttempts; i++ {
			logger.Debug("Creating instance group manager")
			_, err = Sdk.Service.InstanceGroupManagers.Insert(immutable.ProviderConfig().CloudId, expected.(*InstanceGroup).Location, instanceGroupManager).Do()
			if err == nil {
				break
			}

			logger.Debug("Waiting for instance template to be ready.")
			time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
		}

		logger.Success("Created instance group manager [%s]", templateInstance.Name)
	}

	newResource := &InstanceGroup{
		Shared: Shared{
			Name: r.ServerPool.Name,
			//CloudID: id,
		},
		Image:            expected.(*InstanceGroup).Image,
		Size:             expected.(*InstanceGroup).Size,
		Location:         expected.(*InstanceGroup).Location,
		Count:            expected.(*InstanceGroup).Count,
		BootstrapScripts: expected.(*InstanceGroup).BootstrapScripts,
	}

	providerConfig = immutable.ProviderConfig()
	providerConfig.KubernetesAPI.Endpoint = masterIPPublic
	immutable.SetProviderConfig(providerConfig)

	renderedCluster, err := r.immutableRender(newResource, immutable)
	if err != nil {
		return nil, nil, err
	}
	return renderedCluster, newResource, nil
}

// Delete is used to delete the instances on the cloud provider
func (r *InstanceGroup) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instanceGroup.Delete")
	deleteResource := actual.(*InstanceGroup)
	if deleteResource.Name == "" {
		return nil, nil, fmt.Errorf("Unable to delete instance resource without Name [%s]", deleteResource.Name)
	}

	logger.Success("Deleting InstanceGroup manager [%s]", r.ServerPool.Name)
	_, err := Sdk.Service.InstanceGroupManagers.Get(immutable.ProviderConfig().CloudId, immutable.ProviderConfig().Location, strings.ToLower(r.ServerPool.Name)).Do()
	if err == nil {
		_, err := Sdk.Service.InstanceGroupManagers.Delete(immutable.ProviderConfig().CloudId, immutable.ProviderConfig().Location, strings.ToLower(r.ServerPool.Name)).Do()
		if err != nil {
			return nil, nil, err
		}
	}

	_, err = Sdk.Service.InstanceTemplates.Get(immutable.ProviderConfig().CloudId, strings.ToLower(r.ServerPool.Name)).Do()
	if err == nil {
		err := r.retryDeleteInstanceTemplate(immutable)
		if err != nil {
			return nil, nil, err
		}
	}

	// Kubernetes API
	providerConfig := immutable.ProviderConfig()
	providerConfig.KubernetesAPI.Endpoint = ""
	immutable.SetProviderConfig(providerConfig)
	renderedCluster, err := r.immutableRender(actual, immutable)
	if err != nil {
		return nil, nil, err
	}
	return renderedCluster, actual, nil
}

func (r *InstanceGroup) retryDeleteInstanceTemplate(immutable *cluster.Cluster) error {
	for i := 0; i <= DeleteAttempts; i++ {
		_, err := Sdk.Service.InstanceTemplates.Delete(immutable.ProviderConfig().CloudId, strings.ToLower(r.ServerPool.Name)).Do()
		if err != nil {
			logger.Debug("Waiting for InstanceTemplates.Delete to complete...")
			time.Sleep(time.Duration(DeleteSleepSeconds) * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("Timeout deleting instance templates")
}

func (r *InstanceGroup) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("instanceGroup.Render")
	return inaccurateCluster, nil
}
