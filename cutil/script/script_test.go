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

package script

import "testing"

func TestBuildBootstrapScriptHappy(t *testing.T) {
	scripts := []string{
		"vpn/meshbirdMaster.sh",
		"digitalocean_k8s_ubuntu_16.04_master.sh",
	}
	_, err := BuildBootstrapScript(scripts)
	if err != nil {
		t.Fatalf("Unable to get scripts: %v", err)
	}
}

func TestBuildBootstrapScriptSad(t *testing.T) {
	scripts := []string{
		"vpn/meshbirdMaster.s",
		"digitalocean_k8s_ubuntu_16.04_master.s",
	}
	_, err := BuildBootstrapScript(scripts)
	if err == nil {
		t.Fatalf("Merging non existing scripts: %v", err)
	}
}
