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

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/agent"
	"github.com/kubicorn/kubicorn/pkg/local"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	// ClusterAnnotationKubeconfigLocalFile is a cluster API annotation that allows to
	// specify the local file where to store the kubeconfig get from the cluster
	ClusterAnnotationKubeconfigLocalFile = "cluster.alpha.kubicorn.io/kubeconfig-local-file"
)

func GetConfig(existing *cluster.Cluster) error {
	user := existing.SSH.User
	pubKeyPath := local.Expand(existing.SSH.PublicKeyPath)
	if existing.SSH.Port == "" {
		existing.SSH.Port = "22"
	}

	address := fmt.Sprintf("%s:%s", existing.KubernetesAPI.Endpoint, existing.SSH.Port)
	localPath, localPathAnnotationDefined := existing.Annotations[ClusterAnnotationKubeconfigLocalFile]
	if localPathAnnotationDefined {
		localPath = local.Expand(localPath)
	} else {
		var err error
		localDir := filepath.Join(local.Home(), "/.kube")
		localPath, err = getKubeConfigPath(localDir)
		if err != nil {
			return err
		}
	}

	remotePath := ""
	if user == "root" { // --------------------------------------------------------------------------------
		remotePath = "/root/.kube/config"
	} else {
		remotePath = filepath.Join("/home", user, ".kube/config")
	}

	// --------------------------------------------------------------------------------
	//
	// @kris-nova
	//
	// We don't need to check a key first, because SSH can support multiple auth implementations
	// So here we just add BOTH a key based auth, and an arbitrary agent. Please don't touch this
	// without talking to @kris-nova first OR unless something is just completely fucked
	//
	sshAgent := agent.NewAgent()
	sshAgentWithKey, err := sshAgent.AddKey(pubKeyPath)
	if err != nil {
		return fmt.Errorf("Unable to add key: %v", err)
	}
	sshConfig := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConfig.Auth = append(sshConfig.Auth, sshAgent.GetAgent())
	sshConfig.Auth = append(sshConfig.Auth, sshAgentWithKey.GetAgent())
	sshConfig.SetDefaults()
	//
	// --------------------------------------------------------------------------------

	conn, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	c, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer c.Close()
	r, err := c.Open(remotePath)
	if err != nil {
		return err
	}
	defer r.Close()
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) || localPathAnnotationDefined {
		empty := []byte("")
		err := ioutil.WriteFile(localPath, empty, 0755)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(localPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	_, err = f.WriteString(string(bytes))
	if err != nil {
		return err
	}
	defer f.Close()
	logger.Always("Wrote kubeconfig to [%s]", localPath)
	return nil
}

const (
	// RetryAttempts specifies the amount of retries are allowed when getting a file from a server.
	RetryAttempts = 150
	// RetrySleepSeconds specifies the time to sleep after a failed attempt to get a file form a server.
	RetrySleepSeconds = 5
)

func RetryGetConfig(existing *cluster.Cluster) error {
	for i := 0; i <= RetryAttempts; i++ {
		err := GetConfig(existing)
		if err != nil {
			logger.Debug("Waiting for Kubernetes to come up.. [%v]", err)
			time.Sleep(time.Duration(RetrySleepSeconds) * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("Timedout writing kubeconfig")
}

func getKubeConfigPath(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0777); err != nil {
			return "", err
		}
	}
	return filepath.Join(path, "/config"), nil
}
