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

package retry

import (
	"os"
	"time"

	"github.com/kris-nova/kubicorn/cutil/signals"
)

func Retry(f func(), retries int, timeout time.Duration) {
	done := make(chan bool, 1)

loop:
	for i := 0; i < retries; i++ {
		if signals.GetSignalState() != 0 {
			os.Exit(1)
		}
		go func() {
			f()
			done <- true
		}()
		select {
		case <-done:
			break loop
		case <-time.After(timeout):
			break loop
		}
	}
}
