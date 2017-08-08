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

package logger

import (
	"fmt"
	"github.com/fatih/color"
	"strings"
	"time"
)

var (
	Level    = 2
	Color    = true
	TestMode = false
	Fabulous = false
)

func Always(format string, a ...interface{}) {
	if TestMode {
		fmt.Printf(label(format, "✿"), a...)
		return
	}
	if Fabulous {
		fmt.Fprintf(FabulousWriter, label(format, "✿"), a...)
	} else {
		color.Green(label(format, "✿"), a...)
	}
}

func Info(format string, a ...interface{}) {
	if Level >= 3 {
		if TestMode {
			fmt.Printf(label(format, "✔"), a...)
			return
		}
		if Fabulous {
			fmt.Fprintf(FabulousWriter, label(format, "✔"), a...)
		} else if Color {
			color.Cyan(label(format, "✔"), a...)
		} else {
			fmt.Printf(label(format, "✔"), a...)
		}
	}
}

func Debug(format string, a ...interface{}) {
	if Level >= 4 {
		if TestMode {
			fmt.Printf(label(format, "▶"), a...)
			return
		}
		fmt.Printf(label(format, "▶"), a...)
	}
}

func Warning(format string, a ...interface{}) {
	if Level >= 2 {
		if TestMode {
			fmt.Printf(label(format, "!"), a...)
			return
		}
		if Fabulous {
			fmt.Fprintf(FabulousWriter, label(format, "!"), a...)
		} else if Color {
			color.Green(label(format, "!"), a...)
		} else {
			fmt.Printf(label(format, "!"), a...)
		}
	}
}

func Critical(format string, a ...interface{}) {
	if Level >= 1 {
		if TestMode {
			fmt.Printf(label(format, "✖"), a...)
			return
		}
		if Fabulous {
			fmt.Fprintf(FabulousWriter, label(format, "✖"), a...)
		} else if Color {
			color.Red(label(format, "✖"), a...)
		} else {
			fmt.Printf(label(format, "✖"), a...)
		}
	}
}

func Log(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func label(format, label string) string {
	t := time.Now()
	rfct := t.Format(time.RFC3339)
	if !strings.Contains(format, "\n") {
		format = fmt.Sprintf("%s%s", format, "\n")
	}
	return fmt.Sprintf("%s [%s]  %s", rfct, label, format)
}
