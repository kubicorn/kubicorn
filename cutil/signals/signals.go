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

// Package signals exposes signal handler.
package signals

import (
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/kris-nova/kubicorn/cutil/logger"
)

const (
	// signalAbort is used to gracefully exit program.
	signalAbort = 1 << iota
	// signalTerminate is used to terminate program.
	signalTerminate
)

// Signal is an interface that implements signal handling.
type Signal interface {
	GetState() int
	Register()
}

// Handler defines signal handler properties.
type Handler struct {

	// todo (@xmudrii) Can we move these to package level vars instead of in the Handler{}

	// timeoutSeconds defines when handler will timeout in seconds.
	timeoutSeconds int
	// signals stores signals recieved from the system.
	signals chan os.Signal
	// signalReceived is used to store signal handler state.
	signalReceived int
}

// NewSignalHandler creates a new Handler using given properties.
func NewSignalHandler(timeoutSeconds int) *Handler {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)
	return &Handler{
		timeoutSeconds: timeoutSeconds,
		signals:        signals,
		signalReceived: 0,
	}
}

// GetState returns has signal been recieved.
func (h *Handler) GetState() int {
	return h.signalReceived
}

// Register starts handling signals.
func (h *Handler) Register() {
	go func() {
		for {
			select {
			case s := <-h.signals:
				switch {
				case s == os.Interrupt:
					if h.signalReceived == 0 {
						h.signalReceived = 1
						logger.Debug("SIGINT Received")
						continue
					}
					h.signalReceived = signalTerminate
					debug.PrintStack()
					os.Exit(130)
					break
				case s == syscall.SIGQUIT:
					h.signalReceived = signalAbort
					break
				case s == syscall.SIGTERM:
					h.signalReceived = signalTerminate
					os.Exit(3)
					break
				}
			case <-time.After(time.Duration(h.timeoutSeconds) * time.Second):
				os.Exit(4)
				break
			}
		}

	}()
}
