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

package k8slogger

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kris-nova/kubicorn/cutil/logger"
	v1 "k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	// RetryAttempts specifies the amount of retries are allowed when getting a file from a server.
	retryAttempts = 500
	// RetrySleepSeconds specifies the time to sleep after a failed attempt to get a file form a server.
	retrySleepSeconds = 20
)

// GetPodLogsStream return ReadCloser structure with logs from a pod.
func GetPodLogsStream(client *k8s.Clientset, podName, podNamespace string) (io.ReadCloser, error) {
	pods := client.CoreV1().Pods(podNamespace)
	return pods.GetLogs(podName, &v1.PodLogOptions{}).Stream()
}

// PodLogsStreamToString converts ReadCloser to string.
func PodLogsStreamToString(rc io.ReadCloser) string {
	b := new(bytes.Buffer)
	b.ReadFrom(rc)
	return b.String()
}

// WaitPodLogsStream wait for specific log entry provided as a string.
func WaitPodLogsStream(client *k8s.Clientset, podName, podNamespace string) error {
	for i := 0; i <= retryAttempts; i++ {
		rc, err := GetPodLogsStream(client, podName, podNamespace)
		if err != nil {
			logger.Debug("Waiting for Pods to become created.. [%v]", err)
			time.Sleep(time.Duration(retrySleepSeconds) * time.Second)
			continue
		}
		s := PodLogsStreamToString(rc)
		if !strings.Contains(s, "no-exit was specified, sonobuoy is now blocking") {
			logger.Debug("Waiting for Sonobuoy...")
			time.Sleep(time.Duration(retrySleepSeconds) * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("Timedout waiting pods to become created")
}
