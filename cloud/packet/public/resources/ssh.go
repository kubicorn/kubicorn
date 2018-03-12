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
	"os"
	"strings"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/packethost/packngo"
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
			Name: r.Name,
		},
		User:                 immutable.ProviderConfig().SSH.User,
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
	}

	// we need to get the project ID first - because there is no way in kubicorn to pass these things around
	logger.Debug("ssh.Actual finding project ID by name %s", immutable.ProviderConfig().Project.Name)
	project, err := GetProjectByName(immutable.ProviderConfig().Project.Name)
	if err != nil {
		return nil, nil, err
	}
	if project == nil || project.ID == "" {
		newCluster := r.immutableRender(newResource, immutable)
		logger.Debug("ssh.Actual no project, no keys, newResource %v", newResource)
		return newCluster, newResource, nil
	}
	logger.Debug("ssh.Actual project keys %v", project.SSHKeys)
	logger.Debug("ssh.Actual target fingerprint %v", immutable.ProviderConfig().SSH.PublicKeyFingerprint)
	var key packngo.SSHKey
	for _, k := range project.SSHKeys {
		// get the url to check - because Project.SSHKeys does not actually contain the data
		keyID := strings.Replace(k.URL, "/ssh-keys/", "", 1)
		logger.Debug("looking for keyID %v", keyID)
		keyData, response, err := Sdk.Client.SSHKeys.Get(keyID)
		if err != nil && response.StatusCode != 404 {
			return nil, nil, err
		}
		logger.Debug("keyData found '%v'", keyData)
		if keyData != nil && keyData.FingerPrint == immutable.ProviderConfig().SSH.PublicKeyFingerprint {
			key = *keyData
			break
		}
	}

	if key.ID != "" {
		logger.Debug("ssh.Actual found key %v", key)
		newResource.Name = key.Label
		newResource.User = immutable.ProviderConfig().SSH.User
		newResource.PublicKeyData = key.Key
		newResource.PublicKeyFingerprint = key.FingerPrint
		newResource.Identifier = key.ID
	}
	newCluster := r.immutableRender(newResource, immutable)
	logger.Debug("ssh.Actual newResource %v", newResource)
	return newCluster, newResource, nil
}

func (r *SSH) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Expected")
	newResource := &SSH{
		Shared: Shared{
			Name: r.Name,
		},
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		User:                 immutable.ProviderConfig().SSH.User,
	}
	logger.Debug("ssh.Expected newResource %v", newResource)
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *SSH) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Apply")
	expectedResource := expected.(*SSH)
	actualResource := actual.(*SSH)
	logger.Debug("ssh.Apply expectedResource vs actualResource %v %v", *expectedResource, *actualResource)
	if expectedResource.PublicKeyFingerprint == actualResource.PublicKeyFingerprint && actualResource.Identifier != "" {
		logger.Debug("ssh.Apply already equal")
		return immutable, expectedResource, nil
	}

	// if we made it here, we do not have that key, so create it
	request := &packngo.SSHKeyCreateRequest{
		Label:     expectedResource.Name,
		Key:       expectedResource.PublicKeyData,
		ProjectID: immutable.ProviderConfig().Project.Identifier,
	}
	logger.Debug("ssh.Apply creating key %v", request)
	key, _, err := Sdk.Client.SSHKeys.Create(request)
	if err != nil {
		// really should check if key already is in use
		return nil, nil, err
	}
	logger.Success("Created SSH Key [%s]", key.ID)

	newResource := &SSH{
		Shared: Shared{
			Name: r.Name,
		},
		PublicKeyFingerprint: key.FingerPrint,
		PublicKeyData:        key.Key,
		PublicKeyPath:        expectedResource.PublicKeyPath,
		User:                 expectedResource.User,
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *SSH) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Delete")
	actualResource := actual.(*SSH)
	force := strings.ToLower(os.Getenv("KUBICORN_FORCE_DELETE_KEY"))
	logger.Debug("ssh.Delete force %v", force)
	if force == "true" {
		deleteResource := actual.(*SSH)
		if deleteResource.Identifier == "" {
			return nil, nil, fmt.Errorf("Unable to delete SSH key resource without ID [%s]", deleteResource.Name)
		}
		logger.Debug("ssh.Delete deleting key %s", deleteResource.Identifier)
		_, err := Sdk.Client.SSHKeys.Delete(deleteResource.Identifier)
		if err != nil {
			return nil, nil, err
		}
		logger.Success("Deleted SSH Key [%s]", deleteResource.Identifier)
	}

	newResource := &SSH{}
	newResource.Name = actualResource.Name
	newResource.Tags = actualResource.Tags
	newResource.User = actualResource.User
	newResource.PublicKeyPath = actualResource.PublicKeyPath

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
	providerConfig.SSH.Identifier = newResource.(*SSH).Identifier
	providerConfig.SSH.User = newResource.(*SSH).User
	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}
