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
)

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
	KubicornAutoCreatedGroup = "A FABULOUS security group created by Kubicorn for cluster [%s]"
)

func (r *SecurityGroup) Actual(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("securitygroup.Actual")
	if r.CachedActual != nil {
		logger.Debug("Using cached securitygroup [actual]")
		return r.CachedActual, nil
	}
	actual := &SecurityGroup{
		Shared: Shared{
			Name:        r.Name,
			Tags:        make(map[string]string),
			TagResource: r.TagResource,
		},
	}

	if r.Firewall.Identifier != "" {
		input := &ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{&r.Firewall.Identifier},
		}
		output, err := Sdk.Ec2.DescribeSecurityGroups(input)
		if err != nil {
			return nil, err
		}
		lsn := len(output.SecurityGroups)
		if lsn != 1 {
			return nil, fmt.Errorf("Found [%d] Security Groups for ID [%s]", lsn, r.Firewall.Identifier)
		}
		sg := output.SecurityGroups[0]
		for _, rule := range sg.IpPermissions {
			actual.Rules = append(actual.Rules, &Rule{
				IngressFromPort: int(*rule.FromPort),
				IngressToPort:   int(*rule.ToPort),
				IngressSource:   *rule.IpRanges[0].CidrIp,
				IngressProtocol: *rule.IpProtocol,
			})
		}
		for _, tag := range sg.Tags {
			key := *tag.Key
			val := *tag.Value
			actual.Tags[key] = val
		}
		actual.CloudID = *sg.GroupId
		actual.Name = *sg.GroupName
	}
	r.CachedActual = actual
	return actual, nil
}

func (r *SecurityGroup) Expected(known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("securitygroup.Expected")
	if r.CachedExpected != nil {
		logger.Debug("Using cached Security Group [expected]")
		return r.CachedExpected, nil
	}
	expected := &SecurityGroup{
		Shared: Shared{
			Tags: map[string]string{
				"Name":              r.Name,
				"KubernetesCluster": known.Name,
			},
			CloudID:     r.Firewall.Identifier,
			Name:        r.Firewall.Name,
			TagResource: r.TagResource,
		},
	}
	for _, rule := range r.Firewall.Rules {
		expected.Rules = append(expected.Rules, &Rule{
			IngressSource:   rule.IngressSource,
			IngressToPort:   rule.IngressToPort,
			IngressFromPort: rule.IngressFromPort,
			IngressProtocol: rule.IngressProtocol,
		})
	}
	r.CachedExpected = expected
	return expected, nil
}
func (r *SecurityGroup) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("securitygroup.Apply")
	applyResource := expected.(*SecurityGroup)
	isEqual, err := compare.IsEqual(actual.(*SecurityGroup), expected.(*SecurityGroup))
	if err != nil {
		return nil, err
	}
	if isEqual {
		return applyResource, nil
	}

	input := &ec2.CreateSecurityGroupInput{
		GroupName:   &expected.(*SecurityGroup).Name,
		VpcId:       &applyCluster.Network.Identifier,
		Description: S(fmt.Sprintf(KubicornAutoCreatedGroup, applyCluster.Name)),
	}
	output, err := Sdk.Ec2.CreateSecurityGroup(input)
	if err != nil {
		return nil, err
	}
	logger.Info("Created Security Group [%s]", *output.GroupId)

	newResource := &SecurityGroup{}
	newResource.CloudID = *output.GroupId
	newResource.Name = expected.(*SecurityGroup).Name
	for _, expectedRule := range expected.(*SecurityGroup).Rules {
		input := &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:    &newResource.CloudID,
			ToPort:     I64(expectedRule.IngressToPort),
			FromPort:   I64(expectedRule.IngressFromPort),
			CidrIp:     &expectedRule.IngressSource,
			IpProtocol: S(expectedRule.IngressProtocol),
		}
		_, err := Sdk.Ec2.AuthorizeSecurityGroupIngress(input)
		if err != nil {
			return nil, err
		}
		newResource.Rules = append(newResource.Rules, &Rule{
			IngressSource:   expectedRule.IngressSource,
			IngressToPort:   expectedRule.IngressToPort,
			IngressFromPort: expectedRule.IngressFromPort,
		})
	}
	return newResource, nil
}
func (r *SecurityGroup) Delete(actual cloud.Resource, known *cluster.Cluster) (cloud.Resource, error) {
	logger.Debug("securitygroup.Delete")
	deleteResource := actual.(*SecurityGroup)
	if deleteResource.CloudID == "" {
		return nil, fmt.Errorf("Unable to delete Security Group resource without ID [%s]", deleteResource.Name)
	}

	input := &ec2.DeleteSecurityGroupInput{
		GroupId: &actual.(*SecurityGroup).CloudID,
	}
	_, err := Sdk.Ec2.DeleteSecurityGroup(input)
	if err != nil {
		return nil, err
	}
	logger.Info("Deleted Security Group [%s]", actual.(*SecurityGroup).CloudID)

	newResource := &SecurityGroup{}
	newResource.Tags = actual.(*SecurityGroup).Tags
	newResource.Name = actual.(*SecurityGroup).Name

	return newResource, nil
}

func (r *SecurityGroup) Render(renderResource cloud.Resource, renderCluster *cluster.Cluster) (*cluster.Cluster, error) {
	logger.Debug("securitygroup.Render")
	found := false
	for i := 0; i < len(renderCluster.ServerPools); i++ {
		for j := 0; j < len(renderCluster.ServerPools[i].Firewalls); j++ {
			if renderCluster.ServerPools[i].Firewalls[j].Name == renderResource.(*SecurityGroup).Name {
				found = true
				renderCluster.ServerPools[i].Firewalls[j].Identifier = renderResource.(*SecurityGroup).CloudID
				for _, renderRule := range renderResource.(*SecurityGroup).Rules {
					renderCluster.ServerPools[i].Firewalls[j].Rules = append(renderCluster.ServerPools[i].Firewalls[j].Rules, &cluster.Rule{
						IngressSource:   renderRule.IngressSource,
						IngressFromPort: renderRule.IngressFromPort,
						IngressToPort:   renderRule.IngressToPort,
						IngressProtocol: renderRule.IngressProtocol,
					})
				}
			}
		}
	}

	if !found {
		for i := 0; i < len(renderCluster.ServerPools); i++ {
			if renderCluster.ServerPools[i].Name == r.ServerPool.Name {
				found = true
				var rules []*cluster.Rule
				for _, renderRule := range renderResource.(*SecurityGroup).Rules {
					rules = append(rules, &cluster.Rule{
						IngressSource:   renderRule.IngressSource,
						IngressFromPort: renderRule.IngressFromPort,
						IngressToPort:   renderRule.IngressToPort,
						IngressProtocol: renderRule.IngressProtocol,
					})
				}
				renderCluster.ServerPools[i].Firewalls = append(renderCluster.ServerPools[i].Firewalls, &cluster.Firewall{
					Name:       renderResource.(*SecurityGroup).Name,
					Identifier: renderResource.(*SecurityGroup).CloudID,
					Rules:      rules,
				})

			}
		}
	}

	if !found {
		var rules []*cluster.Rule
		for _, renderRule := range renderResource.(*SecurityGroup).Rules {
			rules = append(rules, &cluster.Rule{
				IngressSource:   renderRule.IngressSource,
				IngressFromPort: renderRule.IngressFromPort,
				IngressToPort:   renderRule.IngressToPort,
				IngressProtocol: renderRule.IngressProtocol,
			})
		}
		firewalls := []*cluster.Firewall{
			{
				Name:       renderResource.(*SecurityGroup).Name,
				Identifier: renderResource.(*SecurityGroup).CloudID,
				Rules:      rules,
			},
		}
		renderCluster.ServerPools = append(renderCluster.ServerPools, &cluster.ServerPool{
			Name:       r.ServerPool.Name,
			Identifier: r.ServerPool.Identifier,
			Firewalls:  firewalls,
		})
	}

	return renderCluster, nil
}

func (r *SecurityGroup) Tag(tags map[string]string) error {
	// Todo tag on another resource
	return nil
}
