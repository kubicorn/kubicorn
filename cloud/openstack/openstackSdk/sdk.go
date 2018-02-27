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

package openstackSdk

import (
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
)

// Sdk represents an Openstack SDK.
type Sdk struct {
	Compute       *gophercloud.ServiceClient
	Network       *gophercloud.ServiceClient
	ObjectStorage *gophercloud.ServiceClient
}

// NewSdk constructs a new Openstack SDK for the specified region.
//
// The following environment variable list is looked up in the user
// environment in order to authenticate to the cloud operator:
// OS_AUTH_URL,
// OS_USERNAME,
// OS_USERID,
// OS_PASSWORD,
// OS_TENANT_ID,
// OS_TENANT_NAME,
// OS_DOMAIN_ID,
// OS_DOMAIN_NAME.
//
// Note that only a susbset of these has to be set since most variables derive
// from one another to allow using either names or ids.
func NewSdk(region string) (*Sdk, error) {
	sdk := &Sdk{}
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	// By default, enable re-authenticating when the token expires. This may be
	// useful when the operator policy enforces a short token validity period
	// and you run into a long deployment.
	authOpts.AllowReauth = true

	client, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}

	//----------------------------
	//
	// Openstack Client Resources
	//
	//----------------------------
	endpointOpts := gophercloud.EndpointOpts{
		Region: region,
	}
	// Compute [Nova]
	if sdk.Compute, err = openstack.NewComputeV2(client, endpointOpts); err != nil {
		return nil, err
	}
	// Network [Neutron]
	if sdk.Network, err = openstack.NewNetworkV2(client, endpointOpts); err != nil {
		return nil, err
	}
	// Object Storage [Swift]
	if sdk.ObjectStorage, err = openstack.NewObjectStorageV1(client, endpointOpts); err != nil {
		return nil, err
	}

	return sdk, nil
}
