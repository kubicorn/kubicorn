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
	"github.com/kris-nova/kubicorn/bootstrap"
)

func BuildBootstrapScript(bootstrapScripts []string) ([]byte, error) {
	var userData []byte
	for _, bootstrapScript := range bootstrapScripts {
		scriptData, err := bootstrap.Asset(fmt.Sprintf("bootstrap/%s", bootstrapScript))
		if err != nil {
			return nil, err
		}
		userData = append(userData, scriptData...)
	}

	return userData, nil;
}