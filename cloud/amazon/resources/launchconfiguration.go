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
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/bootstrap"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/logger"
	"github.com/kris-nova/kubicorn/cutil/script"
)

type Lc struct {
	Shared
	InstanceType     string
	Image            string
	ServerPool       *cluster.ServerPool
	BootstrapScripts []string
}

const (
	MasterIPAttempts               = 40
	MasterIPSleepSecondsPerAttempt = 3
)

func (r *Lc) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("lc.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached LC [actual]")
		return r.CachedActual, nil
	}
	actual := &Lc{
		Shared: Shared{
			Name:        r.Name,
			Tags:        make(map[string]string),
			TagResource: r.TagResource,
		},
	}

	if r.ServerPool.Identifier != "" {
		lcInput := &autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{&r.ServerPool.Identifier},
		}
		lcOutput, err := Sdk.ASG.DescribeLaunchConfigurations(lcInput)
		if err != nil {
			return nil, err
		}
		llc := len(lcOutput.LaunchConfigurations)
		if llc != 1 {
			return nil, fmt.Errorf("Found [%d] Launch Configurations for ID [%s]", llc, r.ServerPool.Identifier)
		}
		lc := lcOutput.LaunchConfigurations[0]
		actual.Image = *lc.ImageId
		actual.InstanceType = *lc.InstanceType
		actual.CloudID = *lc.LaunchConfigurationName
		actual.Tags = map[string]string{
			"Name":              r.Name,
			"KubernetesCluster": known.Name,
		}
	} else {
		actual.Image = r.ServerPool.Image
		actual.InstanceType = r.ServerPool.Size
	}
	actual.BootstrapScripts = r.ServerPool.BootstrapScripts
	r.CachedActual = actual
	return actual, nil
}

func (r *Lc) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("asg.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using cached ASG [expected]")
		return r.CachedExpected, nil
	}
	expected := &Lc{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": known.Name,
			},
			CloudID:     r.Name,
			Name:        r.Name,
			TagResource: r.TagResource,
		},
		InstanceType:     r.ServerPool.Size,
		Image:            r.ServerPool.Image,
		BootstrapScripts: r.ServerPool.BootstrapScripts,
	}
	r.CachedExpected = expected
	return expected, nil
}

func (r *Lc) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("lc.Apply")
	applyResource := expected.(*Lc)
	isEqual, err := compare.IsEqual(actual.(*Lc), expected.(*Lc))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}
	var sgs []*string
	found := false
	for _, serverPool := range applyCluster.ServerPools {
		if serverPool.Name == expected.(*Lc).Name {
			for _, firewall := range serverPool.Firewalls {
				sgs = append(sgs, &firewall.Identifier)
			}
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("Unable to lookup serverpool for Launch Configuration %s", r.Name)
	}

	// --- Hack in here for master IP
	privip := ""
	pubip := ""
	if strings.Contains(r.ServerPool.Name, "node") {
		found := false
		logger.Debug("Tag query: [%s] %s", "Name", fmt.Sprintf("%s.master", applyCluster.Name))
		logger.Debug("Tag query: [%s] %s", "KubernetesCluster", applyCluster.Name)
		for i := 0; i < MasterIPAttempts; i++ {
			logger.Debug("Attempting to lookup master IP for node registration..")
			input := &ec2.DescribeInstancesInput{
				Filters: []*ec2.Filter{
					{
						Name:   S("tag:Name"),
						Values: []*string{S(fmt.Sprintf("%s.master", applyCluster.Name))},
					},
					{
						Name:   S("tag:KubernetesCluster"),
						Values: []*string{S(applyCluster.Name)},
					},
				},
			}
			output, err := Sdk.Ec2.DescribeInstances(input)
			if err != nil {
				return nil, err
			}
			lr := len(output.Reservations)
			if lr == 0 {
				logger.Debug("Found %d Reservations, hanging ", lr)
				time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
				continue
			}
			for _, reservation := range output.Reservations {
				for _, instance := range reservation.Instances {
					if instance.PublicIpAddress != nil {
						privip = *instance.PrivateIpAddress
						pubip = *instance.PublicIpAddress
						applyCluster.Values.ItemMap["INJECTEDMASTER"] = fmt.Sprintf("%s:%s", privip, applyCluster.KubernetesAPI.Port)
						applyCluster.KubernetesAPI.Endpoint = pubip
						logger.Info("Found public IP for master: [%s]", pubip)
						found = true
					}
				}
			}
			if found == true {
				break
			}
			time.Sleep(time.Duration(MasterIPSleepSecondsPerAttempt) * time.Second)
		}
		if !found {
			return nil, fmt.Errorf("Unable to find Master IP")
		}
	}

	newResource := &Lc{}
	userData, err := script.BuildBootstrapScript(r.ServerPool.BootstrapScripts)
	if err != nil {
		return nil, err
	}

	//fmt.Println(string(userData))
	applyCluster.Values.ItemMap["INJECTEDPORT"] = applyCluster.KubernetesAPI.Port
	userData, err = bootstrap.Inject(userData, applyCluster.Values.ItemMap)
	if err != nil {
		return nil, err
	}
	//fmt.Println(string(userData))
	b64data := base64.StdEncoding.EncodeToString(userData)
	lcInput := &autoscaling.CreateLaunchConfigurationInput{
		AssociatePublicIpAddress: B(true),
		LaunchConfigurationName:  &expected.(*Lc).Name,
		ImageId:                  &expected.(*Lc).Image,
		InstanceType:             &expected.(*Lc).InstanceType,
		KeyName:                  &applyCluster.SSH.Identifier,
		SecurityGroups:           sgs,
		UserData:                 &b64data,
	}
	_, err = Sdk.ASG.CreateLaunchConfiguration(lcInput)
	if err != nil {
		return nil, err
	}
	logger.Info("Created Launch Configuration [%s]", r.Name)
	newResource.Image = expected.(*Lc).Image
	newResource.InstanceType = expected.(*Lc).InstanceType
	newResource.Name = expected.(*Lc).Name
	newResource.CloudID = expected.(*Lc).Name
	newResource.BootstrapScripts = r.ServerPool.BootstrapScripts
	return newResource, nil
}

func (r *Lc) Delete(actual cloud.Resource, known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("lc.Delete")
	deleteResource := actual.(*Lc)
	if deleteResource.Name == "" {
		return nil, fmt.Errorf("Unable to delete Launch Configuration resource without Name [%s]", deleteResource.Name)
	}
	input := &autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: &actual.(*Lc).Name,
	}
	_, err := Sdk.ASG.DeleteLaunchConfiguration(input)
	if err != nil {
		return nil, err
	}
	logger.Info("Deleted Launch Configuration [%s]", actual.(*Lc).CloudID)

	// Kubernetes API
	known.KubernetesAPI.Endpoint = ""

	newResource := &Lc{}
	newResource.Name = actual.(*Lc).Name
	newResource.Tags = actual.(*Lc).Tags
	newResource.Image = actual.(*Lc).Image
	newResource.InstanceType = actual.(*Lc).InstanceType
	newResource.BootstrapScripts = actual.(*Lc).BootstrapScripts
	return newResource, nil
}

func (r *Lc) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("lc.Render")
	serverPool := &cluster.ServerPool{}
	serverPool.Image = renderResource.(*Lc).Image
	serverPool.Size = renderResource.(*Lc).InstanceType
	serverPool.BootstrapScripts = renderResource.(*Lc).BootstrapScripts
	found := false
	for i := 0; i < len(renderCluster.ServerPools); i++ {
		if renderCluster.ServerPools[i].Name == renderResource.(*Lc).Name {
			renderCluster.ServerPools[i].Image = renderResource.(*Lc).Image
			renderCluster.ServerPools[i].Size = renderResource.(*Lc).InstanceType
			renderCluster.ServerPools[i].BootstrapScripts = renderResource.(*Lc).BootstrapScripts
			found = true
		}
	}
	if !found {
		renderCluster.ServerPools = append(renderCluster.ServerPools, serverPool)
	}

	return renderCluster, nil
}

func (r *Lc) Tag(tags map[string]string) error {
	// Todo tag on another resource
	return nil
}
