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
	"fmt"

	"encoding/json"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/bootstrap"
)

const bootstrapInitScript = "bootstrap_init.sh"
const kubicornDir = "/etc/kubicorn"
const clusterAsJSONFileName = "cluster.json"

func BuildBootstrapScript(bootstrapScripts []string, cluster *cluster.Cluster) ([]byte, error) {
	userData := []byte{}
	if cluster.Cloud == "amazon" {
		scriptData, err := buildBootstrapSetupScript(cluster, kubicornDir, clusterAsJSONFileName)
		if err != nil {
			return nil, err
		}
		userData = append(userData, scriptData...)
	}

	for _, bootstrapScript := range bootstrapScripts {
		scriptData, err := bootstrap.Asset(fmt.Sprintf("bootstrap/%s", bootstrapScript))
		if err != nil {
			return nil, err
		}
		userData = append(userData, scriptData...)
	}

	return userData, nil
}

func buildBootstrapSetupScript(cluster *cluster.Cluster, dir, file string) ([]byte, error) {
	userData, err := bootstrap.Asset(fmt.Sprintf("bootstrap/%s", bootstrapInitScript))
	if err != nil {
		return nil, err
	}
	script := []byte("mkdir -p " + dir + "\nsudo sh -c 'cat <<EOF > " + dir + "/" + file + "\n")

	clusterJSON, err := json.Marshal(cluster)
	if err != nil {
		return nil, err
	}

	userData = append(userData, script...)
	userData = append(userData, clusterJSON...)
	userData = append(userData, []byte("\nEOF'\n")...)
	return userData, nil
}
