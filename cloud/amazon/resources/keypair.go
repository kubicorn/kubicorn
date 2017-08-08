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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"strings"
)

type KeyPair struct {
	Shared
	PublicKeyData        string
	PublicKeyPath        string
	PublicKeyFingerprint string
	User                 string
}

func (r *KeyPair) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("keypair.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached keypair [actual]")
		return r.CachedActual, nil
	}
	actual := &KeyPair{
		Shared: Shared{
			Name:        r.Name,
			Tags:        make(map[string]string),
			TagResource: r.TagResource,
		},
	}

	if known.SSH.Identifier != "" {
		input := &ec2.DescribeKeyPairsInput{
			KeyNames: []*string{&known.SSH.Identifier},
		}
		output, err := Sdk.Ec2.DescribeKeyPairs(input)
		if err == nil {
			lsn := len(output.KeyPairs)
			if lsn != 1 {
				return nil, fmt.Errorf("Found [%d] Keypairs for ID [%s]", lsn, known.Ssh.Identifier)
			}
			keypair := output.KeyPairs[0]
			actual.CloudID = *keypair.KeyName
			actual.PublicKeyFingerprint = *keypair.KeyFingerprint
		}
		lsn := len(output.KeyPairs)
		if lsn != 1 {
			return nil, fmt.Errorf("Found [%d] Keypairs for ID [%s]", lsn, known.SSH.Identifier)
		}
		keypair := output.KeyPairs[0]
		actual.CloudID = *keypair.KeyName
		actual.PublicKeyFingerprint = *keypair.KeyFingerprint
	}
	actual.PublicKeyPath = known.Ssh.PublicKeyPath
	actual.PublicKeyData = string(known.Ssh.PublicKeyData)
	actual.PublicKeyFingerprint = known.Ssh.PublicKeyFingerprint
	actual.User = known.Ssh.User
	r.CachedActual = actual
	return actual, nil
}

func (r *KeyPair) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("keypair.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using keypair subnet [expected]")
		return r.CachedExpected, nil
	}
	expected := &KeyPair{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": known.Name,
			},
			CloudID:     known.SSH.Identifier,
			Name:        r.Name,
			TagResource: r.TagResource,
		},
		PublicKeyPath: known.SSH.PublicKeyPath,
		PublicKeyData: string(known.SSH.PublicKeyData),
		User:          known.SSH.User,
	}
	r.CachedExpected = expected
	return expected, nil
}
func (r *KeyPair) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("keypair.Apply")
	applyResource := expected.(*KeyPair)
	isEqual, err := compare.IsEqual(actual.(*KeyPair), expected.(*KeyPair))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}
	input := &ec2.ImportKeyPairInput{
		KeyName:           &expected.(*KeyPair).Name,
		PublicKeyMaterial: []byte(expected.(*KeyPair).PublicKeyData),
	}
	newResource := &KeyPair{}
	output, err := Sdk.Ec2.ImportKeyPair(input)
	if err != nil {
		if !strings.Contains(err.Error(), "InvalidKeyPair.Duplicate") {
			return nil, err
		}
		logger.Info("Using existing KeyPair [%s]", expected.(*KeyPair).Name)
	} else {
		logger.Info("Created KeyPair [%s]", *output.KeyName)
		newResource.PublicKeyFingerprint = *output.KeyFingerprint
	}
	newResource.CloudID = expected.(*KeyPair).Name
	newResource.PublicKeyData = expected.(*KeyPair).PublicKeyData
	newResource.PublicKeyPath = expected.(*KeyPair).PublicKeyPath
	newResource.User = expected.(*KeyPair).User
	newResource.Name = expected.(*KeyPair).Name
	return newResource, nil
}
func (r *KeyPair) Delete(actual cloud.Resource, known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("keypair.Delete")
	force := false
	if force {
		deleteResource := actual.(*KeyPair)
		if deleteResource.CloudID == "" {
			return nil, fmt.Errorf("Unable to delete keypair resource without ID [%s]", deleteResource.Name)
		}
		input := &ec2.DeleteKeyPairInput{
			KeyName: &actual.(*KeyPair).Name,
		}
		_, err := Sdk.Ec2.DeleteKeyPair(input)
		if err != nil {
			return nil, err
		}
		logger.Info("Deleted keypair [%s]", actual.(*KeyPair).CloudID)
	}
	newResource := &KeyPair{}
	newResource.Tags = actual.(*KeyPair).Tags
	newResource.Name = actual.(*KeyPair).Name
	newResource.PublicKeyPath = actual.(*KeyPair).PublicKeyPath
	return newResource, nil
}

func (r *KeyPair) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("keypair.Render")
	renderCluster.SSH.Name = renderResource.(*KeyPair).Name
	renderCluster.SSH.Identifier = renderResource.(*KeyPair).Name
	renderCluster.SSH.PublicKeyData = []byte(renderResource.(*KeyPair).PublicKeyData)
	renderCluster.SSH.PublicKeyFingerprint = renderResource.(*KeyPair).PublicKeyFingerprint
	renderCluster.SSH.PublicKeyPath = renderResource.(*KeyPair).PublicKeyPath
	renderCluster.SSH.User = renderResource.(*KeyPair).User
	return renderCluster, nil
}

func (r *KeyPair) Tag(tags map[string]string) error {
	// Todo tag on another resource
	return nil
}
