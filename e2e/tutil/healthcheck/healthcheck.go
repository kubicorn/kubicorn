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

package healthcheck

const (
	// RetryAttempts specifies the amount of retries are allowed when getting a file from a server.
	retryAttempts = 150
	// RetrySleepSeconds specifies the time to sleep after a failed attempt to get a file form a server.
	retrySleepSeconds = 5
)
