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
	"os"
	"strings"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
)

var _ cloud.Resource = &KeyPair{}

type KeyPair struct {
	Shared
	PublicKeyData        string
	PublicKeyPath        string
	PublicKeyFingerprint string
	User                 string
}

func (r *KeyPair) Actual(immutable *cluster.Cluster) (actual *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("keypair.Actual")
	newResource := &KeyPair{
		Shared: Shared{
			Name: r.Name,
		},
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
		User:                 immutable.ProviderConfig().SSH.User,
	}

	if immutable.ProviderConfig().SSH.Identifier != "" {
		// Find the keypair by name
		keypair, err := keypairs.Get(Sdk.Compute, immutable.ProviderConfig().SSH.User).Extract()
		if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return nil, nil, err
			}
		} else {
			newResource.PublicKeyFingerprint = keypair.Fingerprint
		}
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) Expected(immutable *cluster.Cluster) (expected *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("keypair.Expected")
	newResource := &KeyPair{
		Shared: Shared{
			Identifier: immutable.ProviderConfig().SSH.Identifier,
			Name:       r.Name,
		},
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
		User:                 immutable.ProviderConfig().SSH.User,
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) Apply(actual cloud.Resource, expected cloud.Resource, immutable *cluster.Cluster) (updatedCluster *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("keypair.Apply")
	keypair := expected.(*KeyPair)
	isEqual, err := compare.IsEqual(actual.(*KeyPair), expected.(*KeyPair))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, keypair, nil
	}
	// Create the keypair
	newResource := new(KeyPair)
	res := keypairs.Create(Sdk.Compute, keypairs.CreateOpts{
		Name:      expected.(*KeyPair).Name,
		PublicKey: expected.(*KeyPair).PublicKeyData,
	})
	output, err := res.Extract()
	if err != nil {
		// If the key is already there, use it.
		if !strings.Contains(err.Error(), "already exists") {
			return nil, nil, err
		}
		logger.Info("Using existing KeyPair [%s]", expected.(*KeyPair).Name)
	} else {
		logger.Success("Created KeyPair [%s]", output.Name)
		newResource.PublicKeyFingerprint = output.Fingerprint
	}
	newResource.Identifier = keypair.Identifier
	newResource.PublicKeyData = keypair.PublicKeyData
	newResource.PublicKeyPath = keypair.PublicKeyPath
	newResource.User = keypair.User
	newResource.Name = keypair.Name

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) Delete(actual cloud.Resource, immutable *cluster.Cluster) (updatedCluster *cluster.Cluster, resource cloud.Resource, err error) {
	logger.Debug("keypair.Delete")
	keypair := actual.(*KeyPair)

	// Delete the keypair
	if strings.ToLower(os.Getenv("KUBICORN_FORCE_DELETE_KEY")) == "true" {
		if res := keypairs.Delete(Sdk.Compute, keypair.Name); res.Err != nil {
			return nil, nil, res.Err
		} else {
			logger.Success("Deleted KeyPair [%s]", keypair.Name)
		}
	}

	newResource := &KeyPair{
		Shared: Shared{
			Name: keypair.Name,
		},
		PublicKeyPath: keypair.PublicKeyPath,
		User:          keypair.User,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("keypair.Render")
	keypair := newResource.(*KeyPair)
	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.SSH.PublicKeyData = []byte(keypair.PublicKeyData)
	providerConfig.SSH.PublicKeyPath = keypair.PublicKeyPath
	providerConfig.SSH.PublicKeyFingerprint = keypair.PublicKeyFingerprint
	providerConfig.SSH.User = keypair.User
	providerConfig.SSH.Name = keypair.Name
	providerConfig.SSH.Identifier = keypair.Identifier
	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}
