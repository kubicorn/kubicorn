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

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	// ClusterAnnotationKubeconfigLocalFile is a cluster API annotation that allows to
	// specify the local file where to store the kubeconfig get from the cluster
	ClusterAnnotationKubeconfigLocalFile = "cluster.alpha.kubicorn.io/kubeconfig-local-file"
)

type configLoader struct {
	existing  *cluster.Cluster
	sshconfig *ssh.ClientConfig
}

func newConfigLoader(existing *cluster.Cluster) (*configLoader, error) {
	pubKeyPath := local.Expand(existing.ProviderConfig().SSH.PublicKeyPath)

	// create an agent, either a connection to the one of the environment of the user or
	// (if there is none) a go-implementation of a simple keyring
	sshAgent := agent.NewAgent()
	sshConfig := &ssh.ClientConfig{
		User:            existing.ProviderConfig().SSH.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConfig.Auth = append(sshConfig.Auth, sshAgent.GetAgent())

	// check if pubkey is already in the agent. if it is already loaded
	// in the agent the user does not need to enter the passphrase
	if err := sshAgent.CheckKey(pubKeyPath); err != nil {
		// add the key to the agent.
		// THIS HAS A SIDEEFFECT: if the agent is the user's agent, the key will
		// be appended to the list of keys. so the agent will have an additional
		// key in it's ring. i don't see a big problem here.
		_, err := sshAgent.AddKey(pubKeyPath)
		if err != nil {
			return nil, fmt.Errorf("Unable to add key: %v", err)
		}
	}
	sshConfig.SetDefaults()

	return &configLoader{
		existing:  existing,
		sshconfig: sshConfig,
	}, nil
}

func getRemoteFileBytes(address, filePath string, sshCfg *ssh.ClientConfig) ([]byte, error) {
	conn, err := ssh.Dial("tcp", address, sshCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	r, err := c.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func getRemoteKubeconfigPath(user string) string {
	if user == "root" {
		return "/root/.kube/config"
	}
	return filepath.Join("/home", user, ".kube/config")
}

func (cfg *configLoader) GetConfig() error {
	user := cfg.existing.ProviderConfig().SSH.User
	if cfg.existing.ProviderConfig().SSH.Port == "" {
		providerConfig := cfg.existing.ProviderConfig()
		providerConfig.SSH.Port = "22"
		cfg.existing.SetProviderConfig(providerConfig)
	}

	address := fmt.Sprintf("%s:%s", cfg.existing.ProviderConfig().KubernetesAPI.Endpoint, cfg.existing.ProviderConfig().SSH.Port)
	localPath, localPathAnnotationDefined := cfg.existing.Annotations[ClusterAnnotationKubeconfigLocalFile]
	if localPathAnnotationDefined {
		localPath = local.Expand(localPath)
	} else {
		var err error
		localPath = GetKubeConfigPath(cfg.existing)
		if err != nil {
			return err
		}
	}

	bytes, err := getRemoteFileBytes(address, getRemoteKubeconfigPath(user), cfg.sshconfig)
	if err != nil {
		return err
	}

	var existingKubeConfig, newKubeConfig *clientcmdapi.Config
	existingKubeConfig = nil
	newKubeConfig, err = clientcmd.Load(bytes)
	if err != nil {
		return err
	}

	if _, err := os.Stat(localPath); os.IsNotExist(err) || localPathAnnotationDefined {
		empty := []byte("")
		err := ioutil.WriteFile(localPath, empty, 0755)
		if err != nil {
			return err
		}
	} else {
		existingKubeConfig, err = clientcmd.LoadFromFile(localPath)
		if err != nil {
			return err
		}
	}

	merged := mergeKubeconfigs([]*clientcmdapi.Config{existingKubeConfig, newKubeConfig})
	merged.CurrentContext = newKubeConfig.CurrentContext
	return clientcmd.WriteToFile(*merged, localPath)
}

const (
	// RetryAttempts specifies the amount of retries are allowed when getting a file from a server.
	RetryAttempts = 150
	// RetrySleepSeconds specifies the time to sleep after a failed attempt to get a file form a server.
	RetrySleepSeconds = 5
)

func GetConfig(existing *cluster.Cluster) error {
	loader, err := newConfigLoader(existing)
	if err != nil {
		return fmt.Errorf("cannot create a config loader: %v", err)
	}
	return loader.GetConfig()
}

func RetryGetConfig(existing *cluster.Cluster) error {
	loader, err := newConfigLoader(existing)
	if err != nil {
		return fmt.Errorf("cannot create a config loader: %v", err)
	}
	for i := 0; i <= RetryAttempts; i++ {
		err := loader.GetConfig()
		if err != nil {
			logger.Debug("Waiting for Kubernetes to come up.. [%v]", err)
			time.Sleep(time.Duration(RetrySleepSeconds) * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("Timedout writing kubeconfig")
}

func getPath(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0777); err != nil {
			return "", err
		}
	}
	return filepath.Join(path, "/config"), nil
}

func mergeKubeconfigs(configs []*clientcmdapi.Config) *clientcmdapi.Config {
	mergedConfig := clientcmdapi.NewConfig()

	for _, config := range configs {
		if config == nil {
			continue
		}
		// merge clusters
		for cName, c := range config.Clusters {
			mergedConfig.Clusters[cName] = c
		}

		// merge authinfos
		for aName, a := range config.AuthInfos {
			mergedConfig.AuthInfos[aName] = a
		}

		// merge contexts
		for ctxName, ctx := range config.Contexts {
			mergedConfig.Contexts[ctxName] = ctx
		}

		// merge extensions
		for extName, ext := range config.Extensions {
			mergedConfig.Extensions[extName] = ext
		}
	}

	return mergedConfig
}

func GetKubeConfigPath(c *cluster.Cluster) string {
	localPath, localPathAnnotationDefined := c.Annotations[ClusterAnnotationKubeconfigLocalFile]
	if localPathAnnotationDefined {
		localPath = local.Expand(localPath)
	} else {
		var err error
		localDir := filepath.Join(local.Home(), "/.kube")
		localPath, err = getPath(localDir)
		if err != nil {
			logger.Warning("Unable to get kubeconfig: %v", err)
			return ""
		}
	}
	return localPath
}
