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
	"context"
	"fmt"
	"strconv"

	"github.com/digitalocean/godo"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &SSH{}

type SSH struct {
	Shared
	User                 string
	PublicKeyFingerprint string
	PublicKeyData        string
	PublicKeyPath        string
}

func (r *SSH) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Actual")
	newResource := &SSH{
		Shared: Shared{
			Name:    r.Name,
			CloudID: immutable.ProviderConfig().SSH.Identifier,
		},
		User: immutable.ProviderConfig().SSH.User,
	}

	if r.CloudID != "" {

		id, err := strconv.Atoi(r.CloudID)
		if err != nil {
			return nil, nil, err
		}
		ssh, _, err := Sdk.Client.Keys.GetByID(context.TODO(), id)
		if err != nil {
			return nil, nil, err
		}
		strid := strconv.Itoa(ssh.ID)
		newResource.Name = ssh.Name
		newResource.CloudID = strid
		newResource.PublicKeyData = ssh.PublicKey
		newResource.PublicKeyFingerprint = ssh.Fingerprint
	} else {
		found := false
		keys, _, err := Sdk.Client.Keys.List(context.TODO(), &godo.ListOptions{})
		if err != nil {
			return nil, nil, err
		}
		for _, key := range keys {
			if key.Fingerprint == immutable.ProviderConfig().SSH.PublicKeyFingerprint {
				found = true
				newResource.Name = key.Name
				newResource.CloudID = strconv.Itoa(key.ID)
				newResource.PublicKeyData = key.PublicKey
				newResource.PublicKeyFingerprint = key.Fingerprint
				newResource.PublicKeyPath = immutable.ProviderConfig().SSH.PublicKeyPath
			}
		}
		if !found {
			newResource.PublicKeyPath = immutable.ProviderConfig().SSH.PublicKeyPath
			newResource.PublicKeyFingerprint = immutable.ProviderConfig().SSH.PublicKeyFingerprint
			newResource.PublicKeyData = string(immutable.ProviderConfig().SSH.PublicKeyData)
			newResource.User = immutable.ProviderConfig().SSH.User
			newResource.CloudID = immutable.ProviderConfig().SSH.Identifier
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *SSH) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Expected")
	newResource := &SSH{
		Shared: Shared{
			Name:    r.Name,
			CloudID: immutable.ProviderConfig().SSH.Identifier,
		},
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		User:                 immutable.ProviderConfig().SSH.User,
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *SSH) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("droplet.Apply")
	applyResource := expected.(*SSH)
	isEqual, err := compare.IsEqual(actual.(*SSH), expected.(*SSH))
	if err != nil {
		return nil, nil, err
	}
	if isEqual && actual.(*SSH).Shared.CloudID != "" {
		return immutable, applyResource, nil
	}
	request := &godo.KeyCreateRequest{
		Name:      expected.(*SSH).Name,
		PublicKey: expected.(*SSH).PublicKeyData,
	}
	key, _, err := Sdk.Client.Keys.Create(context.TODO(), request)
	if err != nil {
		godoErr := err.(*godo.ErrorResponse)
		if godoErr.Message != "SSH Key is already in use on your account" {
			return nil, nil, err
		}
		key, _, err = Sdk.Client.Keys.GetByFingerprint(context.TODO(), expected.(*SSH).PublicKeyFingerprint)
		if err != nil {
			return nil, nil, err
		}
		logger.Info("Using existing SSH Key [%s]", actual.(*SSH).Name)
	} else {
		logger.Success("Created SSH Key [%d]", key.ID)
	}

	id := strconv.Itoa(key.ID)
	newResource := &SSH{
		Shared: Shared{
			Name:    key.Name,
			CloudID: id,
		},
		PublicKeyFingerprint: key.Fingerprint,
		PublicKeyData:        key.PublicKey,
		PublicKeyPath:        expected.(*SSH).PublicKeyPath,
		User:                 expected.(*SSH).User,
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}
func (r *SSH) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Delete")
	force := false
	if force {
		deleteResource := actual.(*SSH)
		if deleteResource.CloudID == "" {
			return nil, nil, fmt.Errorf("Unable to delete ssh resource without Id [%s]", deleteResource.Name)
		}
		id, err := strconv.Atoi(immutable.ProviderConfig().SSH.Identifier)
		if err != nil {
			return nil, nil, err
		}

		_, err = Sdk.Client.Keys.DeleteByID(context.TODO(), id)
		if err != nil {
			return nil, nil, err
		}

		logger.Success("Deleted SSH Key [%d]", id)
	}
	newResource := &SSH{}
	newResource.Name = actual.(*SSH).Name
	newResource.Tags = actual.(*SSH).Tags
	newResource.User = actual.(*SSH).User
	newResource.PublicKeyPath = actual.(*SSH).PublicKeyPath

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *SSH) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("ssh.Render")
	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.SSH.PublicKeyData = []byte(newResource.(*SSH).PublicKeyData)
	providerConfig.SSH.PublicKeyFingerprint = newResource.(*SSH).PublicKeyFingerprint
	providerConfig.SSH.PublicKeyPath = newResource.(*SSH).PublicKeyPath
	providerConfig.SSH.Identifier = newResource.(*SSH).CloudID
	providerConfig.SSH.User = newResource.(*SSH).User
	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}
