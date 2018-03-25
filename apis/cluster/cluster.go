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

package cluster

import (
	"encoding/json"

	"github.com/kubicorn/kubicorn/pkg/logger"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

const (
	CloudAmazon       = "amazon"
	CloudAzure        = "azure"
	CloudGoogle       = "google"
	CloudBaremetal    = "baremetal"
	CloudDigitalOcean = "digitalocean"
	CloudOVH          = "ovh"
	CloudPacket       = "packet"
)

// Cluster is what we use internally in Kubicorn.
// This used to be the internal Kubicorn API but
// we are now migrating to the official cluster API (below)
// This represents a Kubernetes Cluster
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Name is the publically available name of the Cluster
	Name string `json:"name,omitempty"`

	// ClusterAPI is the official Kubernetes cluster API
	// We have this wrapped in a larger struct to support the transition
	// and possibly allow for more values downstream
	ClusterAPI *clusterv1.Cluster `json:"clusterAPI,omitempty"`

	// ControlPlane is the control plane MachineSet
	ControlPlane *clusterv1.MachineSet `json:"controlPlane,omitempty"`

	// MachineSets are a subset of worker machines
	MachineSets []*clusterv1.MachineSet `json:"machineSets,omitempty"`

	// ControllerDeployment is the controller to use with controller profiles.
	// Kubicorn will deploy this resource to the Kubernetes cluster after it is online.
	//
	// Here we only allow this single Deployment to be added to bootstrapping logic.
	// The pattern here says that from an arbitrary deployment, you should be able
	// to bootstrap anything else you could need. We default to the kubicorn controller.
	ControllerDeployment *appsv1beta2.Deployment `json:"controllerDeployment,omitempty"`
}

// ProviderConfig is a convenience method that will attempt
// to return a ControlPlaneProviderConfig for a cluster.
// This is useful for managing the legacy API in a clean way.
// This will ignore errors from json.Unmarshal and will simply
// return an empty config.
func (c *Cluster) ProviderConfig() *ControlPlaneProviderConfig {
	//providerConfig providerConfig
	raw := c.ClusterAPI.Spec.ProviderConfig
	providerConfig := &ControlPlaneProviderConfig{}
	err := json.Unmarshal([]byte(raw), providerConfig)
	if err != nil {
		logger.Critical("Unable to unmarshal provider config: %v", err)
	}
	return providerConfig
}

// SetProviderConfig is a convenience method that will attempt
// to set a provider config on a particular cluster. Just like
// it's counterpart ProviderConfig this makes working with the legacy API much easier.
func (c *Cluster) SetProviderConfig(config *ControlPlaneProviderConfig) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		logger.Critical("Unable to marshal provider config: %v", err)
		return err
	}
	str := string(bytes)
	c.ClusterAPI.Spec.ProviderConfig = str
	return nil
}

// MachineProviderConfigs will return all MachineProviderConfigs for a cluster
func (c *Cluster) MachineProviderConfigs() []*MachineProviderConfig {
	var providerConfigs []*MachineProviderConfig
	for _, machineSet := range c.MachineSets {
		raw := machineSet.Spec.Template.Spec.ProviderConfig
		providerConfig := &MachineProviderConfig{}
		err := json.Unmarshal([]byte(raw), providerConfig)
		if err != nil {
			logger.Critical("Unable to unmarshal provider config: %v", err)
		}
		providerConfigs = append(providerConfigs, providerConfig)
	}
	return providerConfigs
}

// SetMachineProviderConfig will attempt to match a provider config to a machine set
// on the "Name" field. If a match cannot be made we warn and move on.
func (c *Cluster) SetMachineProviderConfigs(providerConfigs []*MachineProviderConfig) {
	for _, providerConfig := range providerConfigs {
		name := providerConfig.ServerPool.Name
		found := false
		for _, machineSet := range c.MachineSets {
			if machineSet.Name == name {
				logger.Debug("Matched machine set to provider config: %s", name)
				bytes, err := json.Marshal(providerConfig)
				if err != nil {
					logger.Critical("unable to marshal machine provider config: %v")
					continue
				}
				str := string(bytes)
				machineSet.Spec.Template.Spec.ProviderConfig = str
				found = true
			}
		}

		// TODO
		// @kris-nova
		// Right now if we have a machine provider config and we can't match it
		// we log a warning and move on. We might want to change this to create
		// the machineSet moving forward..
		if !found {
			logger.Warning("Unable to match provider config to machine set: %s", name)
		}

	}

}

func (c *Cluster) ServerPools() []*ServerPool {
	providerConfigs := c.MachineProviderConfigs()
	var serverPools []*ServerPool
	for _, pc := range providerConfigs {
		serverPools = append(serverPools, pc.ServerPool)
	}
	return serverPools
}

func (c *Cluster) ControlPlaneMachineSet() *clusterv1.MachineSet {

	// TODO @kris-nova this logic won't always work, will need to make this smarter
	for _, ms := range c.MachineSets {
		for _, role := range ms.Spec.Template.Spec.Roles {
			if role == clusterv1.MasterRole {
				return ms
			}
		}
	}
	return nil
}

func (c *Cluster) NewMachineSetsFromProviderConfigs(providerConfigs []*MachineProviderConfig) {
	for _, providerConfig := range providerConfigs {
		name := providerConfig.ServerPool.Name
		for _, machineSet := range c.MachineSets {
			if machineSet.Name == name {
				logger.Info("MachineSet already exists with name: %s", name)
				continue
			}
		}

		logger.Debug("Creating new MachineSet: %v", name)
		bytes, err := json.Marshal(providerConfig)
		if err != nil {
			logger.Critical("unable to marshal machine provider config: %v")
			continue
		}
		str := string(bytes)

		// TODO
		// @kris-nova
		// We probably should a have a function/method for this - this seems like common logic.

		var roles []clusterv1.MachineRole
		if providerConfig.ServerPool.Type == ServerPoolTypeMaster {
			roles = append(roles, clusterv1.MasterRole)
		}
		if providerConfig.ServerPool.Type == ServerPoolTypeNode {
			roles = append(roles, clusterv1.NodeRole)
		}

		// -------------------------------------------------
		//
		// Define a new MachineSet
		//
		machineSet := &clusterv1.MachineSet{
			Spec: clusterv1.MachineSetSpec{
				Template: clusterv1.MachineTemplateSpec{
					Spec: clusterv1.MachineSpec{
						ProviderConfig: str,
						Roles:          roles,
					},
				},
			},
		}
		machineSet.Name = name
		//
		//
		// -------------------------------------------------

		c.MachineSets = append(c.MachineSets, machineSet)
	}
}

// ControlPlaneProviderConfig is the legacy Kubicorn API. We are temporarily storing
// this as a ProviderConfig and these fields will slowly transition
// into fields in the official cluster API
type ControlPlaneProviderConfig struct {
	//Name              string         `json:"name,omitempty"`
	Project *Project `json:"project,omitempty"`
	CloudId string   `json:"cloudId,omitempty"`
	//ServerPools       []*ServerPool  `json:"serverPools,omitempty"`
	Cloud           string         `json:"cloud,omitempty"`
	Location        string         `json:"location,omitempty"`
	SSH             *SSH           `json:"SSH,omitempty"`
	Network         *Network       `json:"network,omitempty"`
	Values          *Values        `json:"values,omitempty"`
	KubernetesAPI   *KubernetesAPI `json:"kubernetesAPI,omitempty"`
	GroupIdentifier string         `json:"groupIdentifier,omitempty"`
	Components      *Components    `json:"components,omitempty"`
}

// ControlPlaneProviderConfig is less exciting ProviderConfig, but
// contains everything defined in a ServerPool
type MachineProviderConfig struct {

	// Name is required as it is how we will match our configs
	// to machineSets later
	Name string

	// The legacy configuration for a MachineSet
	ServerPool *ServerPool
}

// NewCluster will initialize a new Cluster
func NewCluster(name string) *Cluster {
	return &Cluster{
		Name: name,
		ClusterAPI: &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: clusterv1.ClusterSpec{},
		},
		ControlPlane: &clusterv1.MachineSet{},
	}
}
