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

package uuid

import (
	"fmt"
	"math/rand"
	"time"
)

// Generates a time ordered UUID. Top 32b are timestamp bottom 96b are random.
func TimeOrderedUUID() string {
	unixTime := uint32(time.Now().UTC().Unix())
	randomPart1 := rand.Uint32()
	randomPart2 := rand.Uint32()
	randomPart3 := rand.Uint32()
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		unixTime,
		uint16(randomPart1>>16),
		uint16(randomPart1&0xffff),
		uint16(randomPart2>>16),
		uint16(randomPart2&0xffff),
		randomPart3)
}
