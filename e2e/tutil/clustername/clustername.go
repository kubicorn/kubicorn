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

package clustername

import (
	"fmt"

	"github.com/kris-nova/kubicorn/cutil/rand"
)

// GetClusterName returns a cluster name based on a provided
// provider shorthand.
func GetClusterName(providerShorthand string) string {
	return fmt.Sprintf("e2e-%s-%s", providerShorthand, randStringRunes(6))
}

// randStringRunes returns random alphanumeric string with the given length.
func randStringRunes(length int) string {
	var letterRunes = []rune("0123456789abcdef")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.GenerateRandomInt(0, len(letterRunes))]
	}
	return string(b)
}
