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

type Retry interface {
	Work()
}

type Worker struct {
	Function func()
	Retries int
	Timeout time.Duration
	Handler signals.Handler
}

func (w *Worker) Work() {
	done := make(chan bool, 1)

loop:
	for i := 0; i < w.Retries; i++ {
		if w.Handler.GetState() != 0 {
			os.Exit(1)
		}
		go func() {
			w.Function()
			done <- true
		}()
		select {
		case <-done:
			break loop
		case <-time.After(w.Timeout):
			break loop
		}
	}
}
