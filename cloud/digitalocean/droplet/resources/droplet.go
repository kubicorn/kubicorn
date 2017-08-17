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
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"github.com/kris-nova/klone/pkg/local"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/bootstrap"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/cutil/scp"
	"github.com/kris-nova/kubicorn/cutil/script"
)

var _ cloud.Resource = &Droplet{}

type Droplet struct {
	Shared
	Region           string
	Size             string
	Image            string
	Count            int
	SSHFingerprint   string
	BootstrapScripts []string
	ServerPool       *cluster.ServerPool
}

const (
	MasterIPAttempts               = 100
	MasterIPSleepSecondsPerAttempt = 5
	DeleteAttempts                 = 25
	DeleteSleepSecondsPerAttempt   = 3
)

func (r *Droplet) Actual(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("droplet.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached droplet [actual]")
		return known, r.CachedActual, nil
	}
	actual := &Droplet{
		Shared: Shared{
			Name:    r.Name,
			CloudID: r.ServerPool.Identifier,
		},
	}

	droplets, _, err := Sdk.Client.Droplets.ListByTag(context.TODO(), r.Name, &godo.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	ld := len(droplets)
	if ld > 0 {
		actual.Count = len(droplets)

		// Todo (@kris-nova) once we start to test these implementations we really need to work on the droplet logic. Right now we just pick the first one..
		droplet := droplets[0]
		id := strconv.Itoa(droplet.ID)
		actual.Name = droplet.Name
		actual.CloudID = id
		actual.Size = droplet.Size.Slug
		actual.Image = droplet.Image.Slug
		actual.Region = droplet.Region.Slug
	}
	actual.BootstrapScripts = r.ServerPool.BootstrapScripts
	actual.SSHFingerprint = known.SSH.PublicKeyFingerprint
	actual.Name = r.Name
	r.CachedActual = actual
	return known, actual, nil
}

func (r *Droplet) Expected(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("droplet.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using droplet subnet [expected]")
		return known, r.CachedExpected, nil
	}
	expected := &Droplet{
		Shared: Shared{
			Name:    r.Name,
			CloudID: r.ServerPool.Identifier,
		},
		Size:             r.ServerPool.Size,
		Region:           known.Location,
		Image:            r.ServerPool.Image,
		Count:            r.ServerPool.MaxCount,
		SSHFingerprint:   known.SSH.PublicKeyFingerprint,
		BootstrapScripts: r.ServerPool.BootstrapScripts,
	}
	r.CachedExpected = expected
	return known, expected, nil
}

func (r *Droplet) Apply(actual, expected cloud.Resource, expectedCluster *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("droplet.Apply")
	applyResource := expected.(*Droplet)
	isEqual, err := compare.IsEqual(actual.(*Droplet), expected.(*Droplet))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return expectedCluster, applyResource, nil
	}

	userData, err := script.BuildBootstrapScript(r.ServerPool.BootstrapScripts)
	if err != nil {
		return nil, nil, err
	}

	//masterIpPrivate := ""
	masterIPPublic := ""
	if r.ServerPool.Type == cluster.ServerPoolTypeNode {
		found := false
		for i := 0; i < MasterIPAttempts; i++ {
			masterTag := ""
			for _, serverPool := range expectedCluster.ServerPools {
				if serverPool.Type == cluster.ServerPoolTypeMaster {
					masterTag = serverPool.Name
				}
			}
			if masterTag == "" {
				return nil, nil, fmt.Errorf("Unable to find master tag for master IP")
			}
			droplets, _, err := Sdk.Client.Droplets.ListByTag(context.TODO(), masterTag, &godo.ListOptions{})
			if err != nil {
				logger.Debug("Hanging for master IP.. (%v)", err)
				time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
				continue
			}
			ld := len(droplets)
			if ld == 0 {
				logger.Debug("Hanging for master IP..")
				time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
				continue
			}
			if ld > 1 {
				return nil, nil, fmt.Errorf("Found [%d] droplets for tag [%s]", ld, masterTag)
			}
			droplet := droplets[0]
			//masterIpPrivate, err = droplet.PrivateIPv4()
			//if err != nil {
			//	return nil, fmt.Errorf("Unable to detect private IP: %v", err)
			//}
			masterIPPublic, err = droplet.PublicIPv4()
			if err != nil {
				return nil, nil, fmt.Errorf("Unable to detect public IP: %v", err)
			}

			logger.Info("Setting up VPN on Droplets... this could take a little bit longer...")
			pubPath := local.Expand(expectedCluster.SSH.PublicKeyPath)
			privPath := strings.Replace(pubPath, ".pub", "", 1)
			scp := scp.NewSecureCopier(expectedCluster.SSH.User, masterIPPublic, "22", privPath)
			masterVpnIP, err := scp.ReadBytes("/tmp/.ip")
			if err != nil {
				logger.Debug("Hanging for VPN IP.. /tmp/.ip (%v)", err)
				time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
				continue
			}
			masterVpnIPStr := strings.Replace(string(masterVpnIP), "\n", "", -1)
			openvpnConfig, err := scp.ReadBytes("/tmp/clients.conf")
			if err != nil {
				logger.Debug("Hanging for VPN config.. /tmp/clients.ovpn (%v)", err)
				time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
				continue
			}
			openvpnConfigEscaped := strings.Replace(string(openvpnConfig), "\n", "\\n", -1)
			found = true
			expectedCluster.Values.ItemMap["INJECTEDMASTER"] = fmt.Sprintf("%s:%s", masterVpnIPStr, expectedCluster.KubernetesAPI.Port)
			expectedCluster.Values.ItemMap["INJECTEDCONF"] = openvpnConfigEscaped
			break
		}
		if !found {
			return nil, nil, fmt.Errorf("Unable to find Master IP after defined wait")
		}
	}

	expectedCluster.Values.ItemMap["INJECTEDPORT"] = expectedCluster.KubernetesAPI.Port
	userData, err = bootstrap.Inject(userData, expectedCluster.Values.ItemMap)
	if err != nil {
		return nil, nil, err
	}

	sshID, err := strconv.Atoi(expectedCluster.SSH.Identifier)
	if err != nil {
		return nil, nil, err
	}

	var droplet *godo.Droplet
	for j := 0; j < expected.(*Droplet).Count; j++ {
		createRequest := &godo.DropletCreateRequest{
			Name:   fmt.Sprintf("%s-%d", expected.(*Droplet).Name, j),
			Region: expected.(*Droplet).Region,
			Size:   expected.(*Droplet).Size,
			Image: godo.DropletCreateImage{
				Slug: expected.(*Droplet).Image,
			},
			Tags:              []string{expected.(*Droplet).Name},
			PrivateNetworking: true,
			SSHKeys: []godo.DropletCreateSSHKey{
				{
					ID:          sshID,
					Fingerprint: expected.(*Droplet).SSHFingerprint,
				},
			},
			UserData: string(userData),
		}
		droplet, _, err = Sdk.Client.Droplets.Create(context.TODO(), createRequest)
		if err != nil {
			return nil, nil, err
		}
		logger.Info("Created Droplet [%d]", droplet.ID)
	}

	newResource := &Droplet{
		Shared: Shared{
			Name:    r.ServerPool.Name,
			CloudID: strconv.Itoa(droplet.ID),
		},
		Image:            droplet.Image.Slug,
		Size:             droplet.Size.Slug,
		Region:           droplet.Region.Slug,
		Count:            expected.(*Droplet).Count,
		BootstrapScripts: expected.(*Droplet).BootstrapScripts,
	}

	expectedCluster.KubernetesAPI.Endpoint = masterIPPublic

	newResource := &Droplet{}
	newResource.Name = actual.(*Droplet).Name
	newResource.Tags = actual.(*Droplet).Tags
	newResource.Image = actual.(*Droplet).Image
	newResource.Size = actual.(*Droplet).Size
	newResource.Count = actual.(*Droplet).Count
	newResource.Region = actual.(*Droplet).Region
	newResource.BootstrapScripts = actual.(*Droplet).BootstrapScripts

	serverPool := &cluster.ServerPool{}
	serverPool.Type = r.ServerPool.Type
	serverPool.Image = actual.(*Droplet).Image
	serverPool.Size = actual.(*Droplet).Size
	serverPool.Name = actual.(*Droplet).Name
	serverPool.MaxCount = actual.(*Droplet).Count
	serverPool.BootstrapScripts = actual.(*Droplet).BootstrapScripts
	found := false
	for i := 0; i < len(expectedCluster.ServerPools); i++ {
		if expectedCluster.ServerPools[i].Name == actual.(*Droplet).Name {
			expectedCluster.ServerPools[i].Image = actual.(*Droplet).Image
			expectedCluster.ServerPools[i].Size = actual.(*Droplet).Size
			expectedCluster.ServerPools[i].MaxCount = actual.(*Droplet).Count
			expectedCluster.ServerPools[i].BootstrapScripts = actual.(*Droplet).BootstrapScripts
			found = true
		}
	}
	if !found {
		expectedCluster.ServerPools = append(expectedCluster.ServerPools, serverPool)
	}
	expectedCluster.Location = actual.(*Droplet).Region

	return expectedCluster, newResource, nil
}
func (r *Droplet) Delete(actual cloud.Resource, known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("droplet.Delete")
	deleteResource := actual.(*Droplet)
	if deleteResource.Name == "" {
		return nil, nil, fmt.Errorf("Unable to delete droplet resource without Name [%s]", deleteResource.Name)
	}

	droplets, _, err := Sdk.Client.Droplets.ListByTag(context.TODO(), r.Name, &godo.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	if len(droplets) != actual.(*Droplet).Count {
		for i := 0; i < DeleteAttempts; i++ {
			logger.Info("Droplet count mis-match, trying query again")
			time.Sleep(5 * time.Second)
			droplets, _, err = Sdk.Client.Droplets.ListByTag(context.TODO(), r.Name, &godo.ListOptions{})
			if err != nil {
				return nil, nil, err
			}
			if len(droplets) == actual.(*Droplet).Count {
				break
			}
		}
	}

	for _, droplet := range droplets {
		for i := 0; i < DeleteAttempts; i++ {
			if droplet.Status == "new" {
				logger.Debug("Waiting for Droplet creation to finish [%d]...", droplet.ID)
				time.Sleep(DeleteSleepSecondsPerAttempt * time.Second)
			} else {
				break
			}
		}
		_, err = Sdk.Client.Droplets.Delete(context.TODO(), droplet.ID)
		if err != nil {
			return nil, nil, err
		}
		logger.Info("Deleted Droplet [%d]", droplet.ID)
	}

	// Kubernetes API
	known.KubernetesAPI.Endpoint = ""

	newResource := &Droplet{}
	newResource.Name = actual.(*Droplet).Name
	newResource.Tags = actual.(*Droplet).Tags
	newResource.Image = actual.(*Droplet).Image
	newResource.Size = actual.(*Droplet).Size
	newResource.Count = actual.(*Droplet).Count
	newResource.Region = actual.(*Droplet).Region
	newResource.BootstrapScripts = actual.(*Droplet).BootstrapScripts

	serverPool := &cluster.ServerPool{}
	serverPool.Type = r.ServerPool.Type
	serverPool.Image = actual.(*Droplet).Image
	serverPool.Size = actual.(*Droplet).Size
	serverPool.Name = actual.(*Droplet).Name
	serverPool.MaxCount = actual.(*Droplet).Count
	serverPool.BootstrapScripts = actual.(*Droplet).BootstrapScripts
	found := false
	for i := 0; i < len(known.ServerPools); i++ {
		if known.ServerPools[i].Name == actual.(*Droplet).Name {
			known.ServerPools[i].Image = actual.(*Droplet).Image
			known.ServerPools[i].Size = actual.(*Droplet).Size
			known.ServerPools[i].MaxCount = actual.(*Droplet).Count
			known.ServerPools[i].BootstrapScripts = actual.(*Droplet).BootstrapScripts
			found = true
		}
	}
	if !found {
		known.ServerPools = append(known.ServerPools, serverPool)
	}
	known.Location = actual.(*Droplet).Region

	return known, newResource, nil
}
