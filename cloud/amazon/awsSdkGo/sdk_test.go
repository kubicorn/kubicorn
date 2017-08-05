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

package awsSdkGo

import (
	"os"
	"testing"
)

var (
	AwsAccessKey       = os.Getenv("AWS_ACCESS_KEY_ID")
	AwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

func TestMain(m *testing.M) {
	m.Run()
	os.Setenv("AWS_ACCESS_KEY_ID", AwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", AwsSecretAccessKey)
}

func TestSdkHappy(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "123")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "123")
	_, err := NewSdk("us-west-2")
	if err != nil {
		t.Fatalf("Unable to get Amazon SDK: %v", err)
	}
}

func TestSdkSad(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "")
	_, err := NewSdk("us-west-2")
	if err == nil {
		t.Fatalf("Able to get Amazon SDK with empty variables")
	}
}
