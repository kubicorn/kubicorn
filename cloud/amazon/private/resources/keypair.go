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

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &KeyPair{}

type KeyPair struct {
	Shared
	PublicKeyData        string
	PublicKeyPath        string
	PublicKeyFingerprint string
	User                 string
}

func (r *KeyPair) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("keypair.Actual")
	newResource := &KeyPair{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
		User:                 immutable.ProviderConfig().SSH.User,
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
	}

	if immutable.ProviderConfig().SSH.Identifier != "" {
		input := &ec2.DescribeKeyPairsInput{
			KeyNames: []*string{&immutable.ProviderConfig().SSH.Identifier},
		}
		output, err := Sdk.Ec2.DescribeKeyPairs(input)
		if err == nil {
			lsn := len(output.KeyPairs)
			if lsn != 1 {
				return nil, nil, fmt.Errorf("Found [%d] Keypairs for ID [%s]", lsn, immutable.ProviderConfig().SSH.Identifier)
			}
			keypair := output.KeyPairs[0]
			newResource.Identifier = *keypair.KeyName
			newResource.PublicKeyFingerprint = *keypair.KeyFingerprint
		}
		newResource.Tags = map[string]string{
			"Name":              r.Name,
			"KubernetesCluster": immutable.Name,
		}
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil

}

func (r *KeyPair) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("keypair.Expected")
	newResource := &KeyPair{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": immutable.Name,
			},
			Identifier: immutable.ProviderConfig().SSH.Identifier,
			Name:       r.Name,
		},
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		User:                 immutable.ProviderConfig().SSH.User,
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
	}
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("keypair.Apply")
	keypair := expected.(*KeyPair)
	isEqual, err := compare.IsEqual(actual.(*KeyPair), expected.(*KeyPair))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, keypair, nil
	}

	// Query for keypair
	input := &ec2.ImportKeyPairInput{
		KeyName:           &expected.(*KeyPair).Name,
		PublicKeyMaterial: []byte(expected.(*KeyPair).PublicKeyData),
	}
	newResource := &KeyPair{}
	output, err := Sdk.Ec2.ImportKeyPair(input)
	if err != nil {
		if !strings.Contains(err.Error(), "InvalidKeyPair.Duplicate") {
			return nil, nil, err
		}
		logger.Info("Using existing KeyPair [%s]", expected.(*KeyPair).Name)
	} else {
		logger.Success("Created KeyPair [%s]", *output.KeyName)
		newResource.PublicKeyFingerprint = *output.KeyFingerprint
	}
	newResource.Identifier = keypair.Name
	newResource.PublicKeyData = keypair.PublicKeyData
	newResource.PublicKeyPath = keypair.PublicKeyPath
	newResource.User = keypair.User
	newResource.Name = keypair.Name

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("keypair.Delete")
	if strings.ToLower(os.Getenv("KUBICORN_FORCE_DELETE_KEY")) == "true" {
		deleteResource := actual.(*KeyPair)
		if deleteResource.Identifier == "" {
			return nil, nil, fmt.Errorf("Unable to delete keypair resource without ID [%s]", deleteResource.Name)
		}
		input := &ec2.DeleteKeyPairInput{
			KeyName: &actual.(*KeyPair).Name,
		}
		_, err := Sdk.Ec2.DeleteKeyPair(input)
		if err != nil {
			return nil, nil, err
		}
		logger.Success("Deleted keypair [%s]", actual.(*KeyPair).Identifier)
	}
	newResource := &KeyPair{}
	newResource.Tags = actual.(*KeyPair).Tags
	newResource.Name = actual.(*KeyPair).Name
	newResource.PublicKeyPath = actual.(*KeyPair).PublicKeyPath
	newResource.User = actual.(*KeyPair).User

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("keypair.Render")
	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.SSH.Name = newResource.(*KeyPair).Name
	providerConfig.SSH.Identifier = newResource.(*KeyPair).Name
	providerConfig.SSH.PublicKeyData = []byte(newResource.(*KeyPair).PublicKeyData)
	providerConfig.SSH.PublicKeyFingerprint = newResource.(*KeyPair).PublicKeyFingerprint
	providerConfig.SSH.PublicKeyPath = newResource.(*KeyPair).PublicKeyPath
	providerConfig.SSH.User = newResource.(*KeyPair).User
	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}
