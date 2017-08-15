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
	"time"
)

const (
	signalAbort = 1 << iota
	signalTerminate
)

type Signal interface {
	GetState() int
	Register()
}

type Handler struct {
	Timeout time.Duration

	signals        chan os.Signal
	signalReceived int
}

func NewSignalHandler(timeout time.Duration) *Handler {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)

	return &Handler{
		Timeout:        timeout,
		signals:        signals,
		signalReceived: 0,
	}
}

func (h *Handler) GetState() int {
	return h.signalReceived
}

func (h *Handler) Register() {
	for {
		select {
		case s := <-h.signals:
			switch {
			case s == os.Interrupt:
				if h.signalReceived == 0 {
					h.signalReceived = signalAbort
					continue
				}
				h.signalReceived = signalTerminate
				os.Exit(130)
				break
			case s == os.Kill:
				h.signalReceived = signalTerminate
				os.Exit(3)
				break
			case s == syscall.SIGQUIT:
				h.signalReceived = signalAbort
				break
			case s == syscall.SIGTERM:
				h.signalReceived = signalTerminate
				os.Exit(3)
				break
			}
		case <-time.After(h.Timeout):
			os.Exit(4)
			break
		}
	}
}
