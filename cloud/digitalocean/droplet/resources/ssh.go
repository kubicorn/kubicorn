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
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/logger"
)

var _ cloud.Resource = &SSH{}

type SSH struct {
	Shared
	User                 string
	PublicKeyFingerprint string
	PublicKeyData        string
	PublicKeyPath        string
}

func (r *SSH) Actual(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached ssh [actual]")
		return nil, r.CachedActual, nil
	}
	actual := &SSH{
		Shared: Shared{
			Name:    r.Name,
			CloudID: known.SSH.Identifier,
		},
		User: known.SSH.User,
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
		actual.Name = ssh.Name
		actual.CloudID = strid
		actual.PublicKeyData = ssh.PublicKey
		actual.PublicKeyFingerprint = ssh.Fingerprint
	}
	actual.PublicKeyPath = known.SSH.PublicKeyPath
	actual.User = known.SSH.User
	r.CachedActual = actual
	return known, actual, nil
}

func (r *SSH) Expected(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using cached ssh [expected]")
		return known, r.CachedExpected, nil
	}
	expected := &SSH{
		Shared: Shared{
			Name:    r.Name,
			CloudID: known.SSH.Identifier,
		},
		PublicKeyFingerprint: known.SSH.PublicKeyFingerprint,
		PublicKeyData:        string(known.SSH.PublicKeyData),
		PublicKeyPath:        known.SSH.PublicKeyPath,
		User:                 known.SSH.User,
	}
	r.CachedExpected = expected
	return known, expected, nil
}

func (r *SSH) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("droplet.Apply")
	applyResource := expected.(*SSH)
	isEqual, err := compare.IsEqual(actual.(*SSH), expected.(*SSH))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return applyCluster, applyResource, nil
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
		logger.Info("Created SSH Key [%d]", key.ID)
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
	renderedCluster, err := r.render(newResource, applyCluster)
	if err != nil {
		return nil, nil, err
	}
	return renderedCluster, newResource, nil
}
func (r *SSH) Delete(actual cloud.Resource, known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("ssh.Delete")
	force := false
	if force {
		deleteResource := actual.(*SSH)
		if deleteResource.CloudID == "" {
			return nil, nil, fmt.Errorf("Unable to delete ssh resource without Id [%s]", deleteResource.Name)
		}
		id, err := strconv.Atoi(known.SSH.Identifier)
		if err != nil {
			return nil, nil, err
		}

		_, err = Sdk.Client.Keys.DeleteByID(context.TODO(), id)
		if err != nil {
			return nil, nil, err
		}

		logger.Info("Deleted SSH Key [%d]", id)
	}
	newResource := &SSH{}
	newResource.Name = actual.(*SSH).Name
	newResource.Tags = actual.(*SSH).Tags
	newResource.User = actual.(*SSH).User
	newResource.PublicKeyPath = actual.(*SSH).PublicKeyPath
	renderedCluster, err := r.render(newResource, known)
	if err != nil {
		return nil, nil, err
	}
	return renderedCluster, newResource, nil
}

func (r *SSH) render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("ssh.Render")
	renderCluster.SSH.PublicKeyData = []byte(renderResource.(*SSH).PublicKeyData)
	renderCluster.SSH.PublicKeyFingerprint = renderResource.(*SSH).PublicKeyFingerprint
	renderCluster.SSH.PublicKeyPath = renderResource.(*SSH).PublicKeyPath
	renderCluster.SSH.Identifier = renderResource.(*SSH).CloudID
	renderCluster.SSH.User = renderResource.(*SSH).User
	return renderCluster, nil
}
