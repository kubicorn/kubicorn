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

package bootstrap

import (
	"fmt"
	"strings"
)

func Inject(data []byte, values map[string]string) ([]byte, error) {
	strData := string(data)
	for k, v := range values {
		fmt.Printf("Replacing key %v value %v\n", k, v)
		strData = strings.Replace(strData, k, v, 1)
	}
	return []byte(strData), nil
}
