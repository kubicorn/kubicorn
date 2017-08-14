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

package signals

import (
	"os"
	"os/signal"
	"syscall"
)

const (
	signalAbort = 1 << iota
	signalTerminate
)

var signalReceived int

func NewSignalHandler() {
	signalReceived = 0

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)

	go handle(signals)
}

func GetSignalState() int {
	return signalReceived
}

func handle(signals <-chan os.Signal) {
	for {
		select {
		case s := <-signals:
			switch {
			case s == os.Interrupt:
				if signalReceived == 0 {
					signalReceived = signalAbort
					continue
				}
				os.Exit(1)
				break
			case s == os.Kill:
				signalReceived = signalTerminate
				os.Exit(2)
				break
			case s == syscall.SIGQUIT:
				signalReceived = signalAbort
				break
			case s == syscall.SIGTERM:
				signalReceived = signalAbort
				break
			}
		}
	}
}
