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
	"os"
	"testing"
)

var (
	PacketAccessToken = os.Getenv("PACKET_APITOKEN")
)

func TestMain(m *testing.M) {
	m.Run()
	os.Setenv("PACKET_APITOKEN", PacketAccessToken)
}

func TestSdkHappy(t *testing.T) {
	os.Setenv("PACKET_APITOKEN", "123")
	_, err := NewSdk()
	if err != nil {
		t.Fatalf("Unable to get Packet SDK: %v", err)
	}
}

func TestSdkSad(t *testing.T) {
	os.Setenv("PACKET_APITOKEN", "")
	_, err := NewSdk()
	if err == nil {
		t.Fatalf("Able to get Packet SDK with empty variables")
	}
}
