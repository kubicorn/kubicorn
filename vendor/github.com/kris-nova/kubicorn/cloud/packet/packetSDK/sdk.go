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

package packetSDK

import (
	"fmt"
	"os"

	"github.com/packethost/packngo"
)

// Sdk represents the client connection to the cloud provider SDK.
type Sdk struct {
	Client *packngo.Client
}

// NewSdk is used to create a Sdk client to connect to the cloud provider.
func NewSdk() (*Sdk, error) {
	sdk := &Sdk{}
	apiToken := getToken()
	if apiToken == "" {
		return nil, fmt.Errorf("Empty $PACKET_APITOKEN")
	}
	sdk.Client = packngo.NewClient("kubicorn", apiToken, nil)

	return sdk, nil
}

func getToken() string {
	return os.Getenv("PACKET_APITOKEN")
}
