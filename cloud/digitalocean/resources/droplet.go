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

func (r *Droplet) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("droplet.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached droplet [actual]")
		return r.CachedActual, nil
	}
	actual := &Droplet{
		Shared: Shared{
			Name:    r.Name,
			CloudID: r.ServerPool.Identifier,
		},
	}

	droplets, _, err := Sdk.Client.Droplets.ListByTag(context.TODO(), r.Name, &godo.ListOptions{})
	if err != nil {
		return nil, err
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
	return actual, nil
}

func (r *Droplet) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("droplet.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using droplet subnet [expected]")
		return r.CachedExpected, nil
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
	return expected, nil
}

func (r *Droplet) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("droplet.Apply")
	applyResource := expected.(*Droplet)
	isEqual, err := compare.IsEqual(actual.(*Droplet), expected.(*Droplet))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}

	userData, err := script.BuildBootstrapScript(r.ServerPool.BootstrapScripts)
	if err != nil {
		return nil, err
	}

	//masterIpPrivate := ""
	masterIPPublic := ""
	if r.ServerPool.Type == cluster.ServerPoolTypeNode {
		found := false
		for i := 0; i < MasterIPAttempts; i++ {
			masterTag := ""
			for _, serverPool := range applyCluster.ServerPools {
				if serverPool.Type == cluster.ServerPoolTypeMaster {
					masterTag = serverPool.Name
				}
			}
			if masterTag == "" {
				return nil, fmt.Errorf("Unable to find master tag for master IP")
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
				return nil, fmt.Errorf("Found [%d] droplets for tag [%s]", ld, masterTag)
			}
			droplet := droplets[0]
			//masterIpPrivate, err = droplet.PrivateIPv4()
			//if err != nil {
			//	return nil, fmt.Errorf("Unable to detect private IP: %v", err)
			//}
			masterIPPublic, err = droplet.PublicIPv4()
			if err != nil {
				return nil, fmt.Errorf("Unable to detect public IP: %v", err)
			}

			logger.Info("Setting up VPN on Droplets... This could a little bit longer...")
			pubPath := local.Expand(applyCluster.SSH.PublicKeyPath)
			privPath := strings.Replace(pubPath, ".pub", "", 1)
			scp := scp.NewSecureCopier(applyCluster.SSH.User, masterIPPublic, "22", privPath)
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
			applyCluster.Values.ItemMap["INJECTEDMASTER"] = fmt.Sprintf("%s:%s", masterVpnIPStr, applyCluster.KubernetesAPI.Port)
			applyCluster.Values.ItemMap["INJECTEDCONF"] = openvpnConfigEscaped
			break
		}
		if !found {
			return nil, fmt.Errorf("Unable to find Master IP after defined wait")
		}
	}

	applyCluster.Values.ItemMap["INJECTEDPORT"] = applyCluster.KubernetesAPI.Port
	userData, err = bootstrap.Inject(userData, applyCluster.Values.ItemMap)
	if err != nil {
		return nil, err
	}

	sshID, err := strconv.Atoi(applyCluster.SSH.Identifier)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		logger.Info("Created Droplet [%d]", droplet.ID)
	}

	newResource := &Droplet{
		Shared: Shared{
			Name: r.ServerPool.Name,
			//CloudID: id,
		},
		Image:            droplet.Image.Slug,
		Size:             droplet.Size.Slug,
		Region:           droplet.Region.Slug,
		Count:            expected.(*Droplet).Count,
		BootstrapScripts: expected.(*Droplet).BootstrapScripts,
	}
	applyCluster.KubernetesAPI.Endpoint = masterIPPublic
	return newResource, nil
}
func (r *Droplet) Delete(actual cloud.Resource, known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("droplet.Delete")
	deleteResource := actual.(*Droplet)
	if deleteResource.Name == "" {
		return nil, fmt.Errorf("Unable to delete droplet resource without Name [%s]", deleteResource.Name)
	}

	droplets, _, err := Sdk.Client.Droplets.ListByTag(context.TODO(), r.Name, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(droplets) != actual.(*Droplet).Count {
		for i := 0; i < DeleteAttempts; i++ {
			logger.Info("Droplet count mis-match, trying query again")
			time.Sleep(5 * time.Second)
			droplets, _, err = Sdk.Client.Droplets.ListByTag(context.TODO(), r.Name, &godo.ListOptions{})
			if err != nil {
				return nil, err
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
			return nil, err
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
	return newResource, nil
}

func (r *Droplet) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("droplet.Render")

	serverPool := &cluster.ServerPool{}
	serverPool.Type = r.ServerPool.Type
	serverPool.Image = renderResource.(*Droplet).Image
	serverPool.Size = renderResource.(*Droplet).Size
	serverPool.Name = renderResource.(*Droplet).Name
	serverPool.MaxCount = renderResource.(*Droplet).Count
	serverPool.BootstrapScripts = renderResource.(*Droplet).BootstrapScripts
	found := false
	for i := 0; i < len(renderCluster.ServerPools); i++ {
		if renderCluster.ServerPools[i].Name == renderResource.(*Droplet).Name {
			renderCluster.ServerPools[i].Image = renderResource.(*Droplet).Image
			renderCluster.ServerPools[i].Size = renderResource.(*Droplet).Size
			renderCluster.ServerPools[i].MaxCount = renderResource.(*Droplet).Count
			renderCluster.ServerPools[i].BootstrapScripts = renderResource.(*Droplet).BootstrapScripts
			found = true
		}
	}
	if !found {
		renderCluster.ServerPools = append(renderCluster.ServerPools, serverPool)
	}
	renderCluster.Location = renderResource.(*Droplet).Region
	return renderCluster, nil
}

func (r *Droplet) Tag(tags map[string]string) error {
	return nil
}
