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

package pkg

import (
	"fmt"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/cloud"
	"github.com/kubicorn/kubicorn/cloud/amazon/awsSdkGo"
	awspub "github.com/kubicorn/kubicorn/cloud/amazon/public"
	ar "github.com/kubicorn/kubicorn/cloud/amazon/public/resources"
	"github.com/kubicorn/kubicorn/cloud/azure/azureSDK"
	azpub "github.com/kubicorn/kubicorn/cloud/azure/public"
	azr "github.com/kubicorn/kubicorn/cloud/azure/public/resources"
	"github.com/kubicorn/kubicorn/cloud/digitalocean/droplet"
	dr "github.com/kubicorn/kubicorn/cloud/digitalocean/droplet/resources"
	"github.com/kubicorn/kubicorn/cloud/digitalocean/godoSdk"
	"github.com/kubicorn/kubicorn/cloud/google/compute"
	gr "github.com/kubicorn/kubicorn/cloud/google/compute/resources"
	"github.com/kubicorn/kubicorn/cloud/google/googleSDK"
	"github.com/kubicorn/kubicorn/cloud/openstack/openstackSdk"
	osr "github.com/kubicorn/kubicorn/cloud/openstack/operator/generic/resources"
	osovh "github.com/kubicorn/kubicorn/cloud/openstack/operator/ovh"
	"github.com/kubicorn/kubicorn/cloud/packet/packetSDK"
	packetpub "github.com/kubicorn/kubicorn/cloud/packet/public"
	packetr "github.com/kubicorn/kubicorn/cloud/packet/public/resources"
)

// RuntimeParameters contains specific parameters that needs to be passed to each
// cloud provider to satisfy their specific configurations needs at runtime while
// using the Reconciler
type RuntimeParameters struct {
	AwsProfile string
}

// GetReconciler gets the correct Reconciler for the cloud provider currenty used.
func GetReconciler(known *cluster.Cluster, runtimeParameters *RuntimeParameters) (reconciler cloud.Reconciler, err error) {
	switch known.ProviderConfig().Cloud {
	case cluster.CloudGoogle:
		sdk, err := googleSDK.NewSdk()
		if err != nil {
			return nil, err
		}
		gr.Sdk = sdk
		return cloud.NewAtomicReconciler(known, compute.NewGoogleComputeModel(known)), nil
	case cluster.CloudDigitalOcean:
		sdk, err := godoSdk.NewSdk()
		if err != nil {
			return nil, err
		}
		dr.Sdk = sdk
		return cloud.NewAtomicReconciler(known, droplet.NewDigitalOceanDropletModel(known)), nil
	case cluster.CloudAmazon:
		sdk, err := awsSdkGo.NewSdk(known.ProviderConfig().Location, runtimeParameters.AwsProfile)
		if err != nil {
			return nil, err
		}
		ar.Sdk = sdk
		return cloud.NewAtomicReconciler(known, awspub.NewAmazonPublicModel(known)), nil
	case cluster.CloudAzure:
		sdk, err := azureSDK.NewSdk()
		if err != nil {
			return nil, err
		}
		azr.Sdk = sdk
		return cloud.NewAtomicReconciler(known, azpub.NewAzurePublicModel(known)), nil
	case cluster.CloudOVH:
		sdk, err := openstackSdk.NewSdk(known.ProviderConfig().Location)
		if err != nil {
			return nil, err
		}
		osr.Sdk = sdk
		return cloud.NewAtomicReconciler(known, osovh.NewOvhPublicModel(known)), nil
	case cluster.CloudPacket:
		sdk, err := packetSDK.NewSdk()
		if err != nil {
			return nil, err
		}
		packetr.Sdk = sdk
		return cloud.NewAtomicReconciler(known, packetpub.NewPacketPublicModel(known)), nil
	default:
		return nil, fmt.Errorf("Invalid cloud type: %s", known.ProviderConfig().Cloud)
	}
}
