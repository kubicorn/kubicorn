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

package cli

import (
	"os"
	"os/user"

	"github.com/kubicorn/kubicorn/pkg/logger"
)

// ExpandPath returns working directory path
func ExpandPath(path string) string {
	switch path {
	case ".":
		wd, err := os.Getwd()
		if err != nil {
			logger.Critical("Unable to get current working directory: %v", err)
			return ""
		}
		path = wd
	case "~":
		homeVar := os.Getenv("HOME")
		if homeVar == "" {
			homeUser, err := user.Current()
			if err != nil {
				logger.Critical("Unable to use user.Current() for user. Maybe a cross compile issue: %v", err)
				return ""
			}
			path = homeUser.HomeDir
		}
	}

	return path
}
