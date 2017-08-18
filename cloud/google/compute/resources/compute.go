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

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/bootstrap"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/cutil/script"
	compute "google.golang.org/api/compute/v1"
)

var _ cloud.Resource = &Instance{}

// Instance is a representation of the server to be created on the cloud provider.
type Instance struct {
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
)

// Actual is used to build a cluster based on instances on the cloud provider.
func (r *Instance) Actual(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instance.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached instance [actual]")
		return known, r.CachedActual, nil
	}
	actual := &Instance{
		Shared: Shared{
			Name:    r.Name,
			CloudID: r.ServerPool.Identifier,
			Labels: map[string]string{
				"group": r.Name,
			},
		},
	}

	project, err := Sdk.Service.Projects.Get(known.Name).Do()
	if err != nil && project != nil {
		instances, err := Sdk.Service.Instances.List(known.Name, known.Location).Do()
		if err != nil {
			return nil, nil, err
		}

		count := len(instances.Items)
		if count > 0 {
			actual.Count = count

			instance := instances.Items[0]
			actual.Name = instance.Name
			actual.CloudID = string(instance.Id)
			actual.Size = instance.Kind
			actual.Image = r.Image
			actual.Location = instance.Zone
			actual.Labels = instance.Labels
		}
	}

	actual.BootstrapScripts = r.ServerPool.BootstrapScripts
	actual.SSHFingerprint = known.SSH.PublicKeyFingerprint
	actual.Name = r.Name
	r.CachedActual = actual
	return known, actual, nil
}

// Expected is used to build a cluster expected to be on the cloud provider.
func (r *Instance) Expected(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instance.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using instance subnet [expected]")
		return known, r.CachedExpected, nil
	}
	expected := &Instance{
		Shared: Shared{
			Name:    r.Name,
			CloudID: r.ServerPool.Identifier,
			Labels: map[string]string{
				"group": r.Name,
			},
		},
		Size:             r.ServerPool.Size,
		Location:         known.Location,
		Image:            r.ServerPool.Image,
		Count:            r.ServerPool.MaxCount,
		SSHFingerprint:   known.SSH.PublicKeyFingerprint,
		BootstrapScripts: r.ServerPool.BootstrapScripts,
	}
	r.CachedExpected = expected
	return known, expected, nil
}

// Apply is used to create the expected resources on the cloud provider.
func (r *Instance) Apply(actualResource, expectedResource cloud.Resource, expectedCluster *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instance.Apply")
	applyResource := expectedResource.(*Instance)
	isEqual, err := compare.IsEqual(actualResource.(*Instance), expectedResource.(*Instance))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return expectedCluster, applyResource, nil
	}

	scripts, err := script.BuildBootstrapScript(r.ServerPool.BootstrapScripts)
	if err != nil {
		return nil, nil, err
	}

	masterIPPrivate := ""
	masterIPPublic := ""
	if r.ServerPool.Type == cluster.ServerPoolTypeNode {
		found := false
		for i := 0; i < MasterIPAttempts; i++ {
			masterTag := ""
			for _, serverPool := range expectedCluster.ServerPools {
				if serverPool.Type == cluster.ServerPoolTypeMaster {
					masterTag = serverPool.Name + "-0"
				}
			}
			if masterTag == "" {
				return nil, nil, fmt.Errorf("Unable to find master tag.")
			}

			instance, err := Sdk.Service.Instances.Get(expectedCluster.Name, expectedResource.(*Instance).Location, masterTag).Do()
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
			expectedCluster.Values.ItemMap["INJECTEDMASTER"] = fmt.Sprintf("%s:%s", masterIPPrivate, expectedCluster.KubernetesAPI.Port)
			break
		}
		if !found {
			return nil, nil, fmt.Errorf("Unable to find Master IP after defined wait")
		}
	}

	expectedCluster.Values.ItemMap["INJECTEDPORT"] = expectedCluster.KubernetesAPI.Port
	scripts, err = bootstrap.Inject(scripts, expectedCluster.Values.ItemMap)
	finalScripts := string(scripts)
	if err != nil {
		return nil, nil, err
	}

	tags := []string{}
	if expectedCluster.KubernetesAPI.Port == "443" {
		tags = append(tags, "https-server")
	}

	if expectedCluster.KubernetesAPI.Port == "80" {
		tags = append(tags, "http-server")
	}

	prefix := "https://www.googleapis.com/compute/v1/projects/" + expectedCluster.Name
	imageURL := "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/" + expectedResource.(*Instance).Image

	for j := 0; j < expectedResource.(*Instance).Count; j++ {
		sshPublicKeyValue := fmt.Sprintf("%s:%s", expectedCluster.SSH.User, string(expectedCluster.SSH.PublicKeyData))
		instance := &compute.Instance{
			Name:        fmt.Sprintf("%s-%d", expectedResource.(*Instance).Name, j),
			MachineType: prefix + "/zones/" + expectedResource.(*Instance).Location + "/machineTypes/" + expectedResource.(*Instance).Size,
			Disks: []*compute.AttachedDisk{
				{
					AutoDelete: true,
					Boot:       true,
					Type:       "PERSISTENT",
					InitializeParams: &compute.AttachedDiskInitializeParams{
						DiskName:    fmt.Sprintf("disk-%s-%d", expectedResource.(*Instance).Name, j),
						SourceImage: imageURL,
					},
				},
			},
			NetworkInterfaces: []*compute.NetworkInterface{
				&compute.NetworkInterface{
					AccessConfigs: []*compute.AccessConfig{
						&compute.AccessConfig{
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
			Labels: map[string]string{
				"group": expectedResource.(*Instance).Name,
			},
		}

		_, err := Sdk.Service.Instances.Insert(expectedCluster.Name, expectedResource.(*Instance).Location, instance).Do()
		if err != nil {
			return nil, nil, err
		}
		logger.Info("Created instance [%s]", instance.Name)
	}

	newResource := &Instance{
		Shared: Shared{
			Name: r.ServerPool.Name,
			//CloudID: id,
		},
		Image:            expectedResource.(*Instance).Image,
		Size:             expectedResource.(*Instance).Size,
		Location:         expectedResource.(*Instance).Location,
		Count:            expectedResource.(*Instance).Count,
		BootstrapScripts: expectedResource.(*Instance).BootstrapScripts,
	}
	expectedCluster.KubernetesAPI.Endpoint = masterIPPublic

	renderedCluster, err := r.render(newResource, expectedCluster)
	if err != nil {
		return nil, nil, err
	}
	return renderedCluster, newResource, nil
}

// Delete is used to delete the instances on the cloud provider
func (r *Instance) Delete(actual cloud.Resource, known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("instance.Delete")
	deleteResource := actual.(*Instance)
	if deleteResource.Name == "" {
		return nil, nil, fmt.Errorf("Unable to delete instance resource without Name [%s]", deleteResource.Name)
	}

	instances, err := Sdk.Service.Instances.List(known.Name, known.Location).Do()
	if err != nil {
		return nil, nil, err
	}

	for _, instance := range instances.Items {
		if instance.Labels["group"] == actual.(*Instance).Labels["group"] {
			_, err = Sdk.Service.Instances.Delete(known.Name, known.Location, instance.Name).Do()
			if err != nil {
				return nil, nil, err
			}
			logger.Info("Deleted Instance [%s]", instance.Name)
		}
	}

	// Kubernetes API
	known.KubernetesAPI.Endpoint = ""
	renderedCluster, err := r.render(actual, known)
	if err != nil {
		return nil, nil, err
	}
	return renderedCluster, actual, nil
}

func (r *Instance) render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("instance.Render")
	found := false
	for i := 0; i < len(renderCluster.ServerPools); i++ {
		if renderCluster.ServerPools[i].Name == renderResource.(*Instance).Name {
			renderCluster.ServerPools[i].Image = renderResource.(*Instance).Image
			renderCluster.ServerPools[i].Size = renderResource.(*Instance).Size
			renderCluster.ServerPools[i].MaxCount = renderResource.(*Instance).Count
			renderCluster.ServerPools[i].Labels = renderResource.(*Instance).Labels
			renderCluster.ServerPools[i].BootstrapScripts = renderResource.(*Instance).BootstrapScripts
			found = true
		}
	}
	if !found {
		serverPool := &cluster.ServerPool{}
		serverPool.Type = r.ServerPool.Type
		serverPool.Image = renderResource.(*Instance).Image
		serverPool.Size = renderResource.(*Instance).Size
		serverPool.Name = renderResource.(*Instance).Name
		serverPool.MaxCount = renderResource.(*Instance).Count
		serverPool.BootstrapScripts = renderResource.(*Instance).BootstrapScripts
		renderCluster.ServerPools = append(renderCluster.ServerPools, serverPool)
	}
	return renderCluster, nil
}
