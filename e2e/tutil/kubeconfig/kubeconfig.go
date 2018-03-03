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

package kubeconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"path/filepath"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/pkg/agent"
	"github.com/kris-nova/kubicorn/pkg/local"
	"github.com/kris-nova/kubicorn/pkg/logger"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// GetConfigFilePath returns path to Kubeconfig file.
func GetConfigFilePath(existing *cluster.Cluster, sshAgent *agent.Keyring) (string, error) {
	user := existing.SSH.User
	address := fmt.Sprintf("%s:%s", existing.KubernetesAPI.Endpoint, existing.SSH.Port)
	pubKeyPath := local.Expand(existing.SSH.PublicKeyPath)
	if existing.SSH.Port == "" {
		existing.SSH.Port = "22"
	}

	remotePath := ""
	if user == "root" {
		remotePath = "/root/.kube/config"
	} else {
		remotePath = filepath.Join("/home", user, ".kube/config")
	}

	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Check for key
	if err := sshAgent.CheckKey(pubKeyPath); err != nil {
		if keyring, err := sshAgent.AddKey(pubKeyPath); err != nil {
			return "", err
		} else {
			sshAgent = keyring
		}
	}

	if sshAgent != nil && os.Getenv("KUBICORN_FORCE_DISABLE_SSH_AGENT") == "" {
		sshConfig.Auth = append(sshConfig.Auth, sshAgent.GetAgent())
	}

	sshConfig.SetDefaults()
	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	c, err := sftp.NewClient(conn)
	if err != nil {
		return "", err
	}
	defer c.Close()
	r, err := c.Open(remotePath)
	if err != nil {
		return "", err
	}
	defer r.Close()
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	dirPath, err := ioutil.TempDir("", ".kube")
	if err != nil {
		return "", err
	}

	localPath, err := getKubeConfigPath(dirPath)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		empty := []byte("")
		err := ioutil.WriteFile(localPath, empty, 0755)
		if err != nil {
			return "", err
		}
	}

	f, err := os.OpenFile(localPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return "", err
	}
	_, err = f.WriteString(string(bytes))
	if err != nil {
		return "", err
	}
	defer f.Close()
	return localPath, nil
}

const (
	// RetryAttempts specifies the amount of retries are allowed when getting a file from a server.
	RetryAttempts = 150
	// RetrySleepSeconds specifies the time to sleep after a failed attempt to get a file form a server.
	RetrySleepSeconds = 5
)

// RetryGetConfigFilePath trys to get Kubeconfig file path until timeout doesn't occurs.
func RetryGetConfigFilePath(existing *cluster.Cluster, sshAgent *agent.Keyring) (string, error) {
	for i := 0; i <= RetryAttempts; i++ {
		file, err := GetConfigFilePath(existing, sshAgent)
		if err != nil {
			logger.Debug("Waiting for Kubernetes to come up.. [%v]", err)
			time.Sleep(time.Duration(RetrySleepSeconds) * time.Second)
			continue
		}
		return file, nil
	}
	return "", fmt.Errorf("Timedout writing kubeconfig")
}

// getKubeConfigPath create file for Kubeconfig
func getKubeConfigPath(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0777); err != nil {
			return "", err
		}
	}
	return filepath.Join(path, "/config"), nil
}
