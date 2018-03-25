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

// Options represents command options.
type Options struct {
	StateStore     string
	StateStorePath string
	Name           string
	CloudID        string
	Set            string
	AwsProfile     string
	GitRemote      string

	S3AccessKey       string
	S3SecretKey       string
	BucketEndpointURL string
	BucketSSL         bool
	BucketName        string
}

// CRDOptions represents getConfig command options.
type CRDOptions struct {
	Options
}

// DeployControllerOptions represents getConfig command options.
type DeployControllerOptions struct {
	Options
}

// EditOptions represents edit command options.
type EditOptions struct {
	Options
	Editor string
}

// GetConfigOptions represents getConfig command options.
type GetConfigOptions struct {
	Options
}

// CreateOptions represents create command options.
type CreateOptions struct {
	Options
	Profile string
}

// DeleteOptions represents delete command options.
type DeleteOptions struct {
	Options
	Purge bool
}

// ApplyOptions represents apply command options.
type ApplyOptions struct {
	Options
}

// ListOptions represents list command options.
type ListOptions struct {
	Options
	Profile string
}

// ExplainOptions represents explain command options.
type ExplainOptions struct {
	Options
	Profile string
	Output  string
}

// VersionOptions contains fields for version output
type VersionOptions struct {
	Version   string `json:"Version"`
	GitCommit string `json:"GitCommit"`
	BuildDate string `json:"BuildDate"`
	GOVersion string `json:"GOVersion"`
	GOARCH    string `json:"GOARCH"`
	GOOS      string `json:"GOOS"`
}
