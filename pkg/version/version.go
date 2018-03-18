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

package version

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kubicorn/kubicorn/pkg/logger"
)

var versionFile = "/src/github.com/kubicorn/kubicorn/VERSION"

type Version struct {
	Version   string
	GitCommit string
	BuildDate string
	GoVersion string
	GoArch    string
	GoOS      string
}

// GetVersion returns Kubicorn version as a Version struct.
func GetVersion() *Version {
	return &Version{
		Version:   parseVersionFile(),
		GitCommit: getGitCommit(),
		BuildDate: time.Now().UTC().String(),
		GoVersion: runtime.Version(),
		GoArch:    runtime.GOARCH,
		GoOS:      runtime.GOOS,
	}
}

func GetVersionJSONStr() string {
	verBytes, err := json.Marshal(GetVersion())
	if err != nil {
		logger.Critical("Unable to marshal version struct: %v", err)
	}

	return string(verBytes)
}

func parseVersionFile() string {
	// TODO(@xmudrii): this is not going to work once we start releasing binaries.
	path := filepath.Join(os.Getenv("GOPATH") + versionFile)
	vBytes, err := ioutil.ReadFile(path)
	if err != nil {
		// ignore error
		return ""
	}
	return string(vBytes)
}

func getGitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// ignore error
		return ""
	}
	return string(output)
}
