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

package task

import (
	"github.com/kris-nova/kubicorn/cutil/logger"
	"time"
)

type Task func() error

var (
	DefaultTicker = time.NewTicker(200 * time.Millisecond)
)

// annotates a task with a description and a sequence of symbols until the task terminates
func RunAnnotated(task Task, description string, symbol string) error {
	donechan := make(chan bool)
	errchan := make(chan error)

	go func() {
		errchan <- task()
	}()

	LogAnnotation(description, symbol, DefaultTicker, donechan)

	err := <- errchan
	donechan <- true

	return err
}

// logs a description and a sequence of symbols (one for each tick) until a quit is received 
func LogAnnotation(description string, symbol string, ticker *time.Ticker, quit <-chan bool) {
	if description != "" {
		logger.Log(description)
	}

	if symbol != "" {
		go func() {
			for {
				select {
				case <-ticker.C:
					logger.Log(symbol)
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}
}
