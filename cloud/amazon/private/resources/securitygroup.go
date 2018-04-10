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
	"strconv"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &SecurityGroup{}

type Rule struct {
	IngressFromPort int
	IngressToPort   int
	IngressSource   string
	IngressProtocol string
}
type SecurityGroup struct {
	Shared
	Firewall   *cluster.Firewall
	ServerPool *cluster.ServerPool
	Rules      []*Rule
}

const (
	KubicornAutoCreatedGroup = "A fabulous security group created by Kubicorn for cluster [%s]"
)

func (r *SecurityGroup) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("securitygroup.Actual")
	newResource := &SecurityGroup{
		Shared: Shared{
			Name: r.Name,
			Tags: make(map[string]string),
		},
	}

	if r.Firewall.Identifier != "" {
		input := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{&r.Firewall.Identifier},
		}
		output, err := Sdk.Ec2.DescribeSecurityGroups(input)
		if err != nil {
			return nil, nil, err
		}
		lsn := len(output.SecurityGroups)
		if lsn != 1 {
			return nil, nil, fmt.Errorf("Found [%d] Security Groups for ID [%s]", lsn, r.Firewall.Identifier)
		}
		sg := output.SecurityGroups[0]
		for _, rule := range sg.IpPermissions {
			new_rule := &Rule{
				IngressSource:   *rule.IpRanges[0].CidrIp,
				IngressProtocol: *rule.IpProtocol,
			}
			if rule.FromPort != nil {
				new_rule.IngressFromPort = int(*rule.FromPort)
			} else {
				new_rule.IngressFromPort = 0
			}
			if rule.ToPort != nil {
				new_rule.IngressToPort = int(*rule.ToPort)
			} else {
				new_rule.IngressToPort = 65535
			}
			newResource.Rules = append(newResource.Rules, new_rule)
		}
		newResource.Tags = map[string]string{
			"Name":              r.Name,
			"KubernetesCluster": immutable.Name,
		}
		for _, tag := range sg.Tags {
			key := *tag.Key
			val := *tag.Value
			newResource.Tags[key] = val
		}
		newResource.Identifier = *sg.GroupId
		newResource.Name = *sg.GroupName
	} else {
		for _, rule := range r.Firewall.IngressRules {
			inPort, err := strToInt(rule.IngressToPort)
			if err != nil {
				return nil, nil, err
			}
			outPort, err := strToInt(rule.IngressFromPort)
			if err != nil {
				return nil, nil, err
			}
			newResource.Rules = append(newResource.Rules, &Rule{
				IngressSource:   rule.IngressSource,
				IngressToPort:   inPort,
				IngressFromPort: outPort,
				IngressProtocol: rule.IngressProtocol,
			})
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *SecurityGroup) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("securitygroup.Expected")
	newResource := &SecurityGroup{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": immutable.Name,
			},
			Identifier: r.Firewall.Identifier,
			Name:       r.Firewall.Name,
		},
	}

	for _, rule := range r.Firewall.IngressRules {
		inPort, err := strToInt(rule.IngressToPort)
		if err != nil {
			return nil, nil, err
		}
		outPort, err := strToInt(rule.IngressFromPort)
		if err != nil {
			return nil, nil, err
		}
		newResource.Rules = append(newResource.Rules, &Rule{
			IngressSource:   rule.IngressSource,
			IngressToPort:   inPort,
			IngressFromPort: outPort,
			IngressProtocol: rule.IngressProtocol,
		})
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}
func (r *SecurityGroup) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("securitygroup.Apply")
	applyResource := expected.(*SecurityGroup)
	isEqual, err := compare.IsEqual(actual.(*SecurityGroup), expected.(*SecurityGroup))
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, applyResource, nil
	}

	input := &ec2.CreateSecurityGroupInput{
		GroupName:   &expected.(*SecurityGroup).Name,
		VpcId:       &immutable.ProviderConfig().Network.Identifier,
		Description: S(fmt.Sprintf(KubicornAutoCreatedGroup, immutable.Name)),
	}
	output, err := Sdk.Ec2.CreateSecurityGroup(input)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Created Security Group [%s]", *output.GroupId)

	newResource := &SecurityGroup{}
	newResource.Identifier = *output.GroupId
	newResource.Name = expected.(*SecurityGroup).Name
	for _, expectedRule := range expected.(*SecurityGroup).Rules {
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:    &newResource.Identifier,
			ToPort:     I64(expectedRule.IngressToPort),
			FromPort:   I64(expectedRule.IngressFromPort),
			CidrIp:     &expectedRule.IngressSource,
			IpProtocol: S(expectedRule.IngressProtocol),
		}
		_, err := Sdk.Ec2.AuthorizeSecurityGroupIngress(input)
		if err != nil {
			return nil, nil, err
		}
		logger.Success("Created security rule (%s) [ingress] %d %s", expected.(*SecurityGroup).Name, expectedRule.IngressToPort, expectedRule.IngressSource)
		newResource.Rules = append(newResource.Rules, &Rule{
			IngressSource:   expectedRule.IngressSource,
			IngressToPort:   expectedRule.IngressToPort,
			IngressFromPort: expectedRule.IngressFromPort,
			IngressProtocol: expectedRule.IngressProtocol,
		})

	}

	// Tag newly created Security Group
	err = newResource.tag(applyResource.Tags)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to tag new Security Group: %v", err)
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}
func (r *SecurityGroup) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("securitygroup.Delete")
	deleteResource := actual.(*SecurityGroup)
	if deleteResource.Identifier == "" {
		return nil, nil, fmt.Errorf("Unable to delete Security Group resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DeleteSecurityGroupInput{
		GroupId: &actual.(*SecurityGroup).Identifier,
	}
	_, err := Sdk.Ec2.DeleteSecurityGroup(input)
	if err != nil {
		return nil, nil, err
	}
	logger.Success("Deleted Security Group [%s]", actual.(*SecurityGroup).Identifier)

	newResource := &SecurityGroup{}
	newResource.Tags = actual.(*SecurityGroup).Tags
	newResource.Name = actual.(*SecurityGroup).Name

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil

}

func (r *SecurityGroup) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("securitygroup.Render")

	newCluster := inaccurateCluster
	found := false
	machineProviderConfigs := newCluster.MachineProviderConfigs()
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		for j := 0; j < len(machineProviderConfig.ServerPool.Firewalls); j++ {
			firewall := machineProviderConfig.ServerPool.Firewalls[j]
			if firewall.Name == newResource.(*SecurityGroup).Name {
				found = true
				firewall.Identifier = newResource.(*SecurityGroup).Identifier
				var ingressRules []*cluster.IngressRule
				for _, renderRule := range newResource.(*SecurityGroup).Rules {
					ingressRules = append(ingressRules, &cluster.IngressRule{
						IngressSource:   renderRule.IngressSource,
						IngressFromPort: strconv.Itoa(renderRule.IngressFromPort),
						IngressToPort:   strconv.Itoa(renderRule.IngressToPort),
						IngressProtocol: renderRule.IngressProtocol,
					})
				}
				firewall.IngressRules = ingressRules
				machineProviderConfig.ServerPool.Firewalls[j] = firewall
				machineProviderConfigs[i] = machineProviderConfig
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}

	//
	if !found {
		machineProviderConfigs := newCluster.MachineProviderConfigs()
		for i := 0; i < len(machineProviderConfigs); i++ {
			machineProviderConfig := machineProviderConfigs[i]
			if machineProviderConfig.Name == r.ServerPool.Name {
				found = true
				var rules []*cluster.IngressRule
				for _, renderRule := range newResource.(*SecurityGroup).Rules {
					rules = append(rules, &cluster.IngressRule{
						IngressSource:   renderRule.IngressSource,
						IngressFromPort: strconv.Itoa(renderRule.IngressFromPort),
						IngressToPort:   strconv.Itoa(renderRule.IngressToPort),
						IngressProtocol: renderRule.IngressProtocol,
					})
				}
				fw := &cluster.Firewall{
					Name:         newResource.(*SecurityGroup).Name,
					Identifier:   newResource.(*SecurityGroup).Identifier,
					IngressRules: rules,
				}
				machineProviderConfig.ServerPool.Firewalls = []*cluster.Firewall{fw}
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}

	if !found {
		var rules []*cluster.IngressRule
		for _, renderRule := range newResource.(*SecurityGroup).Rules {
			rules = append(rules, &cluster.IngressRule{
				IngressSource:   renderRule.IngressSource,
				IngressFromPort: strconv.Itoa(renderRule.IngressFromPort),
				IngressToPort:   strconv.Itoa(renderRule.IngressToPort),
				IngressProtocol: renderRule.IngressProtocol,
			})
		}
		firewalls := []*cluster.Firewall{
			{
				Name:         newResource.(*SecurityGroup).Name,
				Identifier:   newResource.(*SecurityGroup).Identifier,
				IngressRules: rules,
			},
		}

		providerConfig := []*cluster.MachineProviderConfig{
			{
				ServerPool: &cluster.ServerPool{
					Name:       r.ServerPool.Name,
					Identifier: r.ServerPool.Identifier,
					Firewalls:  firewalls,
				},
			},
		}
		newCluster.NewMachineSetsFromProviderConfigs(providerConfig)
	}

	return newCluster
}

func (r *SecurityGroup) tag(tags map[string]string) error {
	logger.Debug("sg.Tag")
	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{&r.Identifier},
	}
	for key, val := range tags {
		logger.Debug("Registering Security Group tag [%s] %s", key, val)
		tagInput.Tags = append(tagInput.Tags, &ec2.Tag{
			Key:   S("%s", key),
			Value: S("%s", val),
		})
	}
	_, err := Sdk.Ec2.CreateTags(tagInput)
	if err != nil {
		return err
	}
	return nil
}

func strToInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to int err: %#v", err)
	}
	return i, nil
}
