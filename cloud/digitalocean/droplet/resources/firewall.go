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
	"encoding/json"
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/pkg/compare"
	"github.com/kubicorn/kubicorn/pkg/logger"
)

var _ cloud.Resource = &Firewall{}

// Firewall holds all the data for DO firewalls.
// We preserve the same tags as DO apis for json marshal and unmarhsalling data.
type Firewall struct {
	Shared
	InboundRules  []InboundRule  `json:"inbound_rules,omitempty"`
	OutboundRules []OutboundRule `json:"outbound_rules,omitempty"`
	DropletIDs    []int          `json:"droplet_ids,omitempty"`
	Tags          []string       `json:"tags,omitempty"` // Droplet tags
	FirewallID    string         `json:"id,omitempty"`
	Status        string         `json:"status,omitempty"`
	Created       string         `json:"created_at,omitempty"`
	ServerPool    *cluster.ServerPool
}

// InboundRule DO Firewall InboundRule rule.
type InboundRule struct {
	Protocol  string   `json:"protocol,omitempty"`
	PortRange string   `json:"ports,omitempty"`
	Source    *Sources `json:"sources,omitempty"`
}

// OutboundRule DO Firewall outbound rule.
type OutboundRule struct {
	Protocol     string        `json:"protocol,omitempty"`
	PortRange    string        `json:"ports,omitempty"`
	Destinations *Destinations `json:"destinations,omitempty"`
}

// Sources DO Firewall Source parameters.
type Sources struct {
	Addresses        []string `json:"addresses,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	DropletIDs       []int    `json:"droplet_ids,omitempty"`
	LoadBalancerUIDs []string `json:"load_balancer_uids,omitempty"`
}

// Destinations DO Firewall destination  parameters.
type Destinations struct {
	Addresses        []string `json:"addresses,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	DropletIDs       []int    `json:"droplet_ids,omitempty"`
	LoadBalancerUIDs []string `json:"load_balancer_uids,omitempty"`
}

// Actual calls DO firewall Api and returns the actual state of firewall in the cloud.
func (r *Firewall) Actual(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("firewall.Actual")

	newResource := defaultFirewallStruct()
	// Digital Firewalls.Get requires firewall ID, which we will not always have.thats why using List.
	firewalls, _, err := Sdk.Client.Firewalls.List(context.TODO(), &godo.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get firwalls info")
	}
	for _, firewall := range firewalls {
		if firewall.Name == r.Name { // In digitalOcean Firwall names are unique.
			// gotcha get all details from this firewall and populate actual.
			firewallBytes, err := json.Marshal(firewall)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal DO firewall details err: %v", err)
			}
			if err := json.Unmarshal(firewallBytes, newResource); err != nil {
				return nil, nil, fmt.Errorf("failed to unmarhal DO firewall details err: %v", err)
			}
			// hack: DO api doesn't take "0" as portRange, but returns "0" for port range in firewall.List.
			for i := 0; i < len(newResource.OutboundRules); i++ {
				if newResource.OutboundRules[i].PortRange == "0" {
					newResource.OutboundRules[i].PortRange = "all"
				}
			}
			for i := 0; i < len(newResource.InboundRules); i++ {
				if newResource.InboundRules[i].PortRange == "0" {
					newResource.InboundRules[i].PortRange = "all"
				}
			}
		}
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

// Expected returns the Firewall structure of what is Expected.
func (r *Firewall) Expected(immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("firewall.Expected")
	newResource := &Firewall{
		Shared: Shared{
			Name:    r.Name,
			CloudID: r.ServerPool.Identifier,
		},
		InboundRules:  r.InboundRules,
		OutboundRules: r.OutboundRules,
		DropletIDs:    r.DropletIDs,
		Tags:          r.Tags,
		FirewallID:    r.FirewallID,
		Status:        r.Status,
		Created:       r.Created,
	}

	//logger.Info("Expected firewall returned is %+v", immutable)
	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil

}

// Apply will compare the actual and expected firewall config, if needed it will create the firewall.
func (r *Firewall) Apply(actual, expected cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("firewall.Apply")
	expectedResource := expected.(*Firewall)
	actualResource := actual.(*Firewall)

	isEqual, err := compare.IsEqual(actualResource, expectedResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return immutable, expected, nil
	}

	firewallRequest := godo.FirewallRequest{
		Name:          expectedResource.Name,
		InboundRules:  convertInRuleType(expectedResource.InboundRules),
		OutboundRules: convertOutRuleType(expectedResource.OutboundRules),
		DropletIDs:    expectedResource.DropletIDs,
		Tags:          expectedResource.Tags,
	}

	firewall, _, err := Sdk.Client.Firewalls.Create(context.TODO(), &firewallRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create the firewall err: %v", err)
	}
	logger.Success("Created Firewall [%s]", firewall.ID)
	newResource := &Firewall{
		Shared: Shared{
			CloudID: firewall.ID,
			Name:    r.Name,
			Tags:    r.Tags,
		},
		DropletIDs:    r.DropletIDs,
		FirewallID:    firewall.ID,
		InboundRules:  r.InboundRules,
		OutboundRules: r.OutboundRules,
		Created:       r.Created,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func (r *Firewall) immutableRender(newResource cloud.Resource, inaccurateCluster *cluster.Cluster) *cluster.Cluster {
	logger.Debug("firewall.Render")
	newCluster := inaccurateCluster

	found := false
	machineProviderConfigs := newCluster.MachineProviderConfigs()
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		for j := 0; j < len(machineProviderConfig.ServerPool.Firewalls); j++ {
			firewall := newResource.(*Firewall)
			if machineProviderConfig.ServerPool.Firewalls[j].Name == firewall.Name {
				found = true
				machineProviderConfig.ServerPool.Firewalls[j].Name = firewall.Name
				machineProviderConfig.ServerPool.Firewalls[j].Identifier = firewall.CloudID
				machineProviderConfig.ServerPool.Firewalls[j].IngressRules = make([]*cluster.IngressRule, len(firewall.InboundRules))
				for k, renderRule := range firewall.InboundRules {
					machineProviderConfig.ServerPool.Firewalls[j].IngressRules[k] = &cluster.IngressRule{
						IngressProtocol: renderRule.Protocol,
						IngressToPort:   renderRule.PortRange,
						IngressSource:   convertInRuleDest(renderRule),
					}
				}
				machineProviderConfig.ServerPool.Firewalls[j].EgressRules = make([]*cluster.EgressRule, len(firewall.OutboundRules))
				for k, renderRule := range firewall.OutboundRules {
					machineProviderConfig.ServerPool.Firewalls[j].EgressRules[k] = &cluster.EgressRule{
						EgressProtocol:    renderRule.Protocol,
						EgressToPort:      renderRule.PortRange,
						EgressDestination: convertOutRuleDest(renderRule),
					}
				}
				machineProviderConfigs[i] = machineProviderConfig
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}

	if !found {
		machineProviderConfigs := newCluster.MachineProviderConfigs()
		for i := 0; i < len(machineProviderConfigs); i++ {
			machineProviderConfig := machineProviderConfigs[i]
			if machineProviderConfig.Name == r.ServerPool.Name {
				found = true
				var inRules []*cluster.IngressRule
				var egRules []*cluster.EgressRule
				firewall := newResource.(*Firewall)
				for _, renderRule := range firewall.InboundRules {
					inRules = append(inRules, &cluster.IngressRule{
						IngressProtocol: renderRule.Protocol,
						IngressToPort:   renderRule.PortRange,
						IngressSource:   convertInRuleDest(renderRule),
					})
				}
				for _, renderRule := range firewall.OutboundRules {
					egRules = append(egRules, &cluster.EgressRule{
						EgressProtocol:    renderRule.Protocol,
						EgressToPort:      renderRule.PortRange,
						EgressDestination: convertOutRuleDest(renderRule),
					})
				}
				machineProviderConfig.ServerPool.Firewalls = append(newCluster.ServerPools()[i].Firewalls, &cluster.Firewall{
					Name:         firewall.Name,
					Identifier:   firewall.CloudID,
					IngressRules: inRules,
					EgressRules:  egRules,
				})
				machineProviderConfigs[i] = machineProviderConfig
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}
	if !found {
		var inRules []*cluster.IngressRule
		var egRules []*cluster.EgressRule
		firewall := newResource.(*Firewall)
		for _, renderRule := range firewall.InboundRules {
			inRules = append(inRules, &cluster.IngressRule{
				IngressProtocol: renderRule.Protocol,
				IngressToPort:   renderRule.PortRange,
				IngressSource:   convertInRuleDest(renderRule),
			})
		}
		for _, renderRule := range firewall.OutboundRules {
			egRules = append(egRules, &cluster.EgressRule{
				EgressProtocol:    renderRule.Protocol,
				EgressToPort:      renderRule.PortRange,
				EgressDestination: convertOutRuleDest(renderRule),
			})
		}
		firewalls := []*cluster.Firewall{
			{
				Name:         firewall.Name,
				Identifier:   firewall.CloudID,
				IngressRules: inRules,
				EgressRules:  egRules,
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

	// Todo (@kris-nova) Figure out what is setting empty firewalls and fix the original bug
	for i := 0; i < len(machineProviderConfigs); i++ {
		machineProviderConfig := machineProviderConfigs[i]
		for j := 0; j < len(machineProviderConfig.ServerPool.Firewalls); j++ {
			firewall := newResource.(*Firewall)
			if firewall.Name == "" {
				logger.Debug("Found empty firewill, will not save!")
				machineProviderConfig.ServerPool.Firewalls = append(machineProviderConfig.ServerPool.Firewalls[:j], machineProviderConfig.ServerPool.Firewalls[j+1:]...)
				machineProviderConfigs[i] = machineProviderConfig
				newCluster.SetMachineProviderConfigs(machineProviderConfigs)
			}
		}
	}

	return newCluster
}

// Delete removes the firewall
func (r *Firewall) Delete(actual cloud.Resource, immutable *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("firewall.Delete")
	deleteResource, ok := actual.(*Firewall)
	if !ok {
		return nil, nil, fmt.Errorf("failed to type convert actual Firewall type ")
	}
	if deleteResource.Name == "" {
		return immutable, nil, nil
		return nil, nil, fmt.Errorf("Unable to delete firewall resource without Name [%s]", deleteResource.Name)
	}
	if _, err := Sdk.Client.Firewalls.Delete(context.TODO(), deleteResource.FirewallID); err != nil {
		return nil, nil, fmt.Errorf("failed to delete firewall [%s] err: %v", deleteResource.Name, err)
	}
	logger.Success("Deleted firewall [%s]", deleteResource.FirewallID)

	newResource := &Firewall{
		Shared: Shared{
			Name: r.Name,
			Tags: r.Tags,
		},
		InboundRules:  r.InboundRules,
		OutboundRules: r.OutboundRules,
		Created:       r.Created,
	}

	newCluster := r.immutableRender(newResource, immutable)
	return newCluster, newResource, nil
}

func defaultFirewallStruct() *Firewall {
	return &Firewall{
		DropletIDs:    make([]int, 0),
		Tags:          make([]string, 0),
		InboundRules:  make([]InboundRule, 0),
		OutboundRules: make([]OutboundRule, 0),
	}
}

func convertInRuleType(rules []InboundRule) []godo.InboundRule {
	inRule := make([]godo.InboundRule, 0)
	for _, rule := range rules {
		source := godo.Sources(*rule.Source)
		godoRule := godo.InboundRule{
			Protocol:  rule.Protocol,
			PortRange: rule.PortRange,
			Sources:   &source,
		}
		inRule = append(inRule, godoRule)
	}
	return inRule
}
func convertOutRuleType(rules []OutboundRule) []godo.OutboundRule {
	outRule := make([]godo.OutboundRule, 0)
	for _, rule := range rules {
		destination := godo.Destinations(*rule.Destinations)
		godoRule := godo.OutboundRule{
			Protocol:     rule.Protocol,
			PortRange:    rule.PortRange,
			Destinations: &destination,
		}
		outRule = append(outRule, godoRule)
	}
	return outRule
}

func convertInRuleDest(src InboundRule) string {
	if len(src.Source.Tags) > 0 && src.Source.Tags[0] != "" {
		return src.Source.Tags[0]
	}
	return src.Source.Addresses[0]
}

func convertOutRuleDest(dest OutboundRule) string {
	if len(dest.Destinations.Tags) > 0 && dest.Destinations.Tags[0] != "" {
		return dest.Destinations.Tags[0]
	}
	return dest.Destinations.Addresses[0]
}
