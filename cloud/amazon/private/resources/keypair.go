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

	"github.com/aws/aws-sdk-go/aws/awserr"
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
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		User:                 immutable.ProviderConfig().SSH.User,
	}

	if immutable.ProviderConfig().SSH.Identifier != "" {
		output, err := Sdk.Ec2.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{
			KeyNames: []*string{&immutable.ProviderConfig().SSH.Identifier},
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				switch awsErr.Code() {
				case "InvalidKeyPair.NotFound":
				default:
					return nil, nil, err
				}
			}
		} else {
			keypair := output.KeyPairs[0]

			newResource.Identifier = *keypair.KeyName
			newResource.PublicKeyFingerprint = *keypair.KeyFingerprint
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("keypair.Expected")

	newResource := &KeyPair{
		Shared: Shared{
			Identifier: r.Name,
			Name:       r.Name,
			Tags:       make(map[string]string),
		},
		PublicKeyData:        string(immutable.ProviderConfig().SSH.PublicKeyData),
		PublicKeyFingerprint: immutable.ProviderConfig().SSH.PublicKeyFingerprint,
		PublicKeyPath:        immutable.ProviderConfig().SSH.PublicKeyPath,
		User:                 immutable.ProviderConfig().SSH.User,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("keypair.Apply")

	applyResource := expected.(*KeyPair)
	isEqual, err := compare.IsEqual(actual.(*KeyPair), applyResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	output, err := Sdk.Ec2.ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           &applyResource.Name,
		PublicKeyMaterial: []byte(applyResource.PublicKeyData),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to import new Key Pair: %v", err)
	}
	logger.Success("Created Key Pair [%s]", *output.KeyName)

	newResource := &KeyPair{
		Shared: Shared{
			Identifier: *output.KeyName,
			Name:       applyResource.Name,
			Tags:       make(map[string]string),
		},
		PublicKeyData:        applyResource.PublicKeyData,
		PublicKeyFingerprint: *output.KeyFingerprint,
		PublicKeyPath:        applyResource.PublicKeyPath,
		User:                 applyResource.User,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("keypair.Delete")

	deleteResource := actual.(*KeyPair)
	if strings.ToLower(os.Getenv("KUBICORN_FORCE_DELETE_KEY")) == "true" {
		if deleteResource.Identifier == "" {
			return nil, nil, fmt.Errorf("Unable to delete Key Pair resource without ID [%s]", deleteResource.Name)
		}
		_, err := Sdk.Ec2.DeleteKeyPair(&ec2.DeleteKeyPairInput{
			KeyName: &deleteResource.Identifier,
		})
		if err != nil {
			return nil, nil, err
		}
		logger.Success("Deleted Key Pair [%s]", deleteResource.Identifier)
	}

	newResource := &KeyPair{
		Shared: Shared{
			Name: deleteResource.Name,
		},
		PublicKeyPath: deleteResource.PublicKeyPath,
		User:          deleteResource.User,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *KeyPair) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("keypair.Render")

	newCluster := inaccurateCluster
	providerConfig := newCluster.ProviderConfig()
	providerConfig.SSH.Identifier = newResource.(*KeyPair).Identifier
	providerConfig.SSH.Name = newResource.(*KeyPair).Name
	providerConfig.SSH.PublicKeyData = []byte(newResource.(*KeyPair).PublicKeyData)
	providerConfig.SSH.PublicKeyFingerprint = newResource.(*KeyPair).PublicKeyFingerprint
	providerConfig.SSH.PublicKeyPath = newResource.(*KeyPair).PublicKeyPath
	providerConfig.SSH.User = newResource.(*KeyPair).User

	newCluster.SetProviderConfig(providerConfig)
	return newCluster
}
