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

package fileresource

import (
	"net/url"
	"os"
	"strings"

	"github.com/kubicorn/kubicorn/pkg/logger"
)

// ReadFromResource reads a file from different sources
// at the moment suppoted resources are http, http, local file system(POSIX)
func ReadFromResource(r string) (string, error) {
	env := os.Getenv("KUBICORN_ENVIRONMENT")

	// Hack in here for local bootstrap override
	devMode := os.Getenv("KUBICORN_FORCE_LOCAL_BOOTSTRAP")
	if devMode != "" {
		logger.Info("Parsing bootstrap script from filesystem [%s]", r)
		return readFromFS(r)
	}

	switch {

	// -----------------------------------------------------------------------------------------------------------------
	//
	//
	// starts with bootstrap/
	//
	case strings.HasPrefix(strings.ToLower(r), "bootstrap/") && env != "LOCAL":

		// If we start with bootstrap/ we know this is a resource we should pull from github.com
		// So here we build the GitHub URL and send the request
		gitHubUrl := getGitHubUrl(r)
		logger.Info("Parsing bootstrap script from GitHub [%s]", gitHubUrl)
		url, err := url.ParseRequestURI(gitHubUrl)
		if err != nil {
			return "", err
		}
		return readFromHTTP(url)

	// -----------------------------------------------------------------------------------------------------------------
	//
	//
	// starts with http(s)://
	//
	case strings.HasPrefix(strings.ToLower(r), "http://") || strings.HasPrefix(strings.ToLower(r), "https://") && env != "LOCAL":
		url, err := url.ParseRequestURI(r)
		logger.Info("Parsing bootstrap script from url [%s]", url)
		if err != nil {
			return "", err
		}
		return readFromHTTP(url)

	// -----------------------------------------------------------------------------------------------------------------
	//
	//
	// pull from local
	//
	default:
		logger.Info("Parsing bootstrap script from filesystem [%s]", r)
		return readFromFS(r)
	}
}
