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

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/profiles/amazon"
)

func TestBuildBootstrapScriptHappy(t *testing.T) {
	testCluster := cluster.NewCluster("testCluster")
	config := &cluster.ControlPlaneProviderConfig{
		Cloud: "digtialocean",
	}
	_ = testCluster.SetProviderConfig(config)
	scripts := []string{
		"bootstrap/digitalocean_k8s_ubuntu_16.04_master.sh",
	}
	_, err := BuildBootstrapScript(scripts, testCluster)
	if err != nil {
		t.Fatalf("Unable to get scripts: %v", err)
	}
}

func TestBuildBootstrapScriptSad(t *testing.T) {
	testCluster := cluster.NewCluster("testCluster")
	config := &cluster.ControlPlaneProviderConfig{
		Cloud: "digtialocean",
	}
	_ = testCluster.SetProviderConfig(config)
	scripts := []string{
		"bootstrap/digitalocean_k8s_ubuntu_16.04_master.s",
	}
	_, err := BuildBootstrapScript(scripts, testCluster)
	if err == nil {
		t.Fatalf("Merging non existing scripts: %v", err)
	}
}

func TestAWSBuildBootstrapScriptLargerThan16KbDoesNotThrowError(t *testing.T) {
	testCluster := cluster.NewCluster("testCluster")
	config := &cluster.ControlPlaneProviderConfig{
		Cloud: "amazon",
	}
	_ = testCluster.SetProviderConfig(config)
	file, e := os.Create("20kb_test_script.sh")
	if e != nil {
		t.Errorf("Could not generate test script")
	}
	if e := file.Truncate(2e4); e != nil {
		t.Errorf("Could not generate test script")
	}
	scripts := []string{
		"20kb_test_script.sh",
	}
	_, err := BuildBootstrapScript(scripts, testCluster)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	_ = os.Remove("20kb_test_script.sh")
}

func TestBuildBootstrapSetupScript(t *testing.T) {
	dir := "."
	fileName := "test.json"
	expectedJsonSetup := `mkdir -p .
cat <<"EOF" > ./test.json`
	expectedEnd := "\nEOF\n"

	c := amazon.NewCentosCluster("bootstrap-setup-script-test")
	os.Remove(dir + "/" + fileName)
	os.Remove("test.sh")
	script, err := buildBootstrapSetupScript(c, dir, fileName)
	if err != nil {
		t.Fatalf("Error building bootstrap setup script: %v", err)
	}
	stringScript := string(script)
	jsonCluster, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("Error marshaling cluster to json: %v", err)
	}
	if shebang := "#!/usr/bin/env bash"; !strings.HasPrefix(stringScript, shebang) {
		t.Fatalf("Expected start of script is wrong!\n\nActual:\n%v\n\nExpected:\n%v", stringScript, shebang)
	}
	if !strings.HasSuffix(stringScript, expectedEnd) {
		t.Fatalf("Expected end of script is wrong!\n\nActual:\n%v\n\nExpected:\n%v", stringScript, expectedEnd)
	}
	if !strings.Contains(stringScript, expectedJsonSetup) {
		t.Fatalf("Expected script to have mkdir followed by writing to file!\n\nActual:\n%v\n\nExpected:\n%v", stringScript, expectedJsonSetup)
	}
	if !strings.Contains(stringScript, string(jsonCluster)) {
		t.Fatal("Json cluster isn't in script!")
	}
}
