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
	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cutil/compare"
	"github.com/kris-nova/kubicorn/cutil/logger"
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
func (f *Firewall) Actual(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Info("Firewall Actual [%s]", f.Name)
	if cached := f.getCachedActual(); cached != nil {
		logger.Debug("Using cached firewall [actual]")
		return known, cached, nil
	}

	actualFirewall := defaultFirewallStruct()
	// Digital Firewalls.Get requires firewall ID, which we will not always have.thats why using List.
	firewalls, _, err := Sdk.Client.Firewalls.List(context.TODO(), &godo.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get firwalls info")
	}
	for _, firewall := range firewalls {
		if firewall.Name == f.Name { // In digitalOcean Firwall names are unique.
			// gotcha get all details from this firewall and populate actual.
			firewallBytes, err := json.Marshal(firewall)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal DO firewall details err: %v", err)
			}
			if err := json.Unmarshal(firewallBytes, actualFirewall); err != nil {
				return nil, nil, fmt.Errorf("failed to unmarhal DO firewall details err: %v", err)
			}
			// hack: DO api doesn't take "0" as portRange, but returns "0" for port range in firewall.List.
			for i := 0; i < len(actualFirewall.OutboundRules); i++ {
				if actualFirewall.OutboundRules[i].PortRange == "0" {
					actualFirewall.OutboundRules[i].PortRange = "all"
				}
			}
			logger.Info("Actual firewall returned is %+v", actualFirewall)
			return known, actualFirewall, nil
		}
	}
	return known, &Firewall{}, nil
}

// Expected returns the Firewall structure of what is Expected.
func (f *Firewall) Expected(known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {

	logger.Info("Firewall Expected [%s]", f.Name)
	if cached := f.getCachedExpected(); cached != nil {
		logger.Debug("Using Expected cached firewall [%s]", f.Name)
		return known, cached, nil
	}
	expected := &Firewall{
		Shared: Shared{
			Name:    f.Name,
			CloudID: f.ServerPool.Identifier,
		},

		InboundRules:  f.InboundRules,
		OutboundRules: f.OutboundRules,
		DropletIDs:    f.DropletIDs,
		Tags:          f.Tags,
		FirewallID:    f.FirewallID,
		Status:        f.Status,
		Created:       f.Created,
	}
	f.CachedExpected = expected
	logger.Info("Expected firewall returned is %+v", expected)
	return known, expected, nil

}

// Apply will compare the actual and expected firewall config, if needed it will create the firewall.
func (f *Firewall) Apply(actual, expected cloud.Resource, applyCluster *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("Firewall.Apply")
	expectedResource, ok := expected.(*Firewall)
	if !ok {
		return nil, nil, fmt.Errorf("Failed to type convert expected Firewall type ")
	}
	actualResource, ok := actual.(*Firewall)
	if !ok {
		return nil, nil, fmt.Errorf("Failed to type convert actual Firewall type ")
	}

	isEqual, err := compare.IsEqual(actualResource, expectedResource)
	if err != nil {
		return nil, nil, err
	}
	if isEqual {
		return nil, expected, nil
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
	f.FirewallID = firewall.ID

	applyCluster, err = f.render(applyCluster, f)
	if err != nil {
		return nil, nil, err
	}

	return applyCluster, f, nil
}

func (f *Firewall) render(renderCluster *cluster.Cluster, renderResource cloud.Resource) (*cluster.Cluster, error) {
	logger.Debug("Firewall.Render")

	found := false
	for i := 0; i < len(renderCluster.ServerPools); i++ {
		for j := 0; j < len(renderCluster.ServerPools[i].Firewalls); j++ {
			firewall, ok := renderResource.(*Firewall)

			if !ok {
				return nil, fmt.Errorf("failed type convert renderResource Firewall type")
			}
			if renderCluster.ServerPools[i].Firewalls[j].Name == firewall.Name {
				found = true
				renderCluster.ServerPools[i].Firewalls[j].Name = firewall.Name
				renderCluster.ServerPools[i].Firewalls[j].Identifier = firewall.CloudID
				renderCluster.ServerPools[i].Firewalls[j].IngressRules = make([]*cluster.IngressRule, len(firewall.InboundRules))
				for k, renderRule := range firewall.InboundRules {
					renderCluster.ServerPools[i].Firewalls[j].IngressRules[k] = &cluster.IngressRule{
						IngressProtocol: renderRule.Protocol,
						IngressToPort:   renderRule.PortRange,
						IngressSource:   renderRule.Source.Addresses[0],
					}
				}
				renderCluster.ServerPools[i].Firewalls[j].EgressRules = make([]*cluster.EgressRule, len(firewall.OutboundRules))
				for k, renderRule := range firewall.OutboundRules {
					renderCluster.ServerPools[i].Firewalls[j].EgressRules[k] = &cluster.EgressRule{
						EgressProtocol:    renderRule.Protocol,
						EgressToPort:      renderRule.PortRange,
						EgressDestination: renderRule.Destinations.Addresses[0],
					}
				}
			}
		}
	}

	if !found {
		for i := 0; i < len(renderCluster.ServerPools); i++ {
			if renderCluster.ServerPools[i].Name == f.ServerPool.Name {
				found = true
				var inRules []*cluster.IngressRule
				var egRules []*cluster.EgressRule
				firewall, ok := renderResource.(*Firewall)
				if !ok {
					return nil, fmt.Errorf("failed type convert renderResource Firewall type")
				}
				for _, renderRule := range firewall.InboundRules {
					inRules = append(inRules, &cluster.IngressRule{
						IngressProtocol: renderRule.Protocol,
						IngressToPort:   renderRule.PortRange,
						IngressSource:   renderRule.Source.Addresses[0],
					})
				}
				for _, renderRule := range firewall.OutboundRules {
					egRules = append(egRules, &cluster.EgressRule{
						EgressProtocol:    renderRule.Protocol,
						EgressToPort:      renderRule.PortRange,
						EgressDestination: renderRule.Destinations.Addresses[0],
					})
				}
				renderCluster.ServerPools[i].Firewalls = append(renderCluster.ServerPools[i].Firewalls, &cluster.Firewall{
					Name:         firewall.Name,
					Identifier:   firewall.CloudID,
					IngressRules: inRules,
					EgressRules:  egRules,
				})
			}
		}
	}
	if !found {
		var inRules []*cluster.IngressRule
		var egRules []*cluster.EgressRule
		firewall, ok := renderResource.(*Firewall)
		if !ok {
			return nil, fmt.Errorf("failed type convert renderResource Firewall type")
		}
		for _, renderRule := range firewall.InboundRules {
			inRules = append(inRules, &cluster.IngressRule{
				IngressProtocol: renderRule.Protocol,
				IngressToPort:   renderRule.PortRange,
				IngressSource:   renderRule.Source.Addresses[0],
			})
		}
		for _, renderRule := range firewall.OutboundRules {
			egRules = append(egRules, &cluster.EgressRule{
				EgressProtocol:    renderRule.Protocol,
				EgressToPort:      renderRule.PortRange,
				EgressDestination: renderRule.Destinations.Addresses[0],
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
		renderCluster.ServerPools = append(renderCluster.ServerPools, &cluster.ServerPool{
			Name:       f.ServerPool.Name,
			Identifier: f.ServerPool.Identifier,
			Firewalls:  firewalls,
		})
	}

	return renderCluster, nil
}

// Delete removes the firewall
func (f *Firewall) Delete(actual cloud.Resource, known *cluster.Cluster) (*cluster.Cluster, cloud.Resource, error) {
	logger.Debug("firewall.Delete")
	deleteResource, ok := actual.(*Firewall)
	if !ok {
		return nil, nil, fmt.Errorf("failed to type convert actual Firewall type ")
	}
	if deleteResource.Name == "" {
		return nil, nil, fmt.Errorf("Unable to delete droplet resource without Name [%s]", deleteResource.Name)
	}
	if _, err := Sdk.Client.Firewalls.Delete(context.TODO(), deleteResource.FirewallID); err != nil {
		return nil, nil, fmt.Errorf("failed to delete firewall [%s] err: %v", deleteResource.Name, err)
	}
	var err error
	known, err = f.render(known, actual)
	if err != nil {
		return nil, nil, err
	}
	return known, actual, nil
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
