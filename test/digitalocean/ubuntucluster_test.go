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

package digitalocean

import (
	"testing"
)

func TestMain(m *testing.M) {
	/*logger.TestMode = true
	logger.Level = 4
	ssh.InitRsaTravis()

	// Create Cluster...
	cluster := digitalocean.NewUbuntuCluster("myCluster")
	cluster, err := initapi.InitCluster(cluster)
	if err != nil {
		panic(err.Error())
	}
	reconciler, err := cutil.GetReconciler(cluster, nil)
	if err != nil {
		panic(err.Error())
	}
	expected, err := reconciler.Expected(cluster)
	if err != nil {
		panic(err.Error())
	}
	actual, err := reconciler.Actual(cluster)
	if err != nil {
		panic(err.Error())
	}
	created, err := reconciler.Reconcile(actual, expected)
	logger.Success("Created cluster [%s]", created.Name)
	if err != nil {
		panic(err.Error())
	}

	sdk, err := godoSdk.NewSdk()
	if err != nil {
		panic(err)
	}
	droplets, _, err := sdk.Client.Droplets.ListByTag(context.TODO(), "myCluster-master", &godo.ListOptions{})
	if err != nil {
		panic(err)
	}

	droplet := droplets[0]
	masterPublicIP, err := droplet.PublicIPv4()
	if err != nil {
		panic(err)
	}

	agent := agent.NewAgent()

	copier := scp.NewSecureCopier("root", masterPublicIP, "22", "/home/marko/.ssh/id_rsa", agent)

	session := command.NewSSHDetails("root", masterPublicIP, "22", "/home/marko/.ssh/id_rsa", agent)
	out, err := session.ExecuteCommand("ls")
	if err != nil {
		panic(err)
	}
	logger.Success("%s", string(out))
	panic(string(out))

	*/
}
