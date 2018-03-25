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

package crd

import (
	"fmt"
	//"io"
	//"io/ioutil"
	//"os"
	//"path"
	//"strings"

	"github.com/ghodss/yaml"
	"github.com/kubicorn/kubicorn/apis/cluster"
	//"github.com/kubicorn/kubicorn/pkg/logger"
	//"github.com/kubicorn/kubicorn/pkg/state"
	"time"

	"strings"

	"github.com/kubicorn/kubicorn/pkg/kubeconfig"
	"github.com/kubicorn/kubicorn/pkg/logger"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/client"
	"k8s.io/kube-deploy/cluster-api/util"
)

const (
	RetryAttempts          = 10
	SleepSecondsPerAttempt = 1
)

type CRDStoreOptions struct {
	ClusterName string
	BasePath    string
}

type CRDStore struct {
	options      *CRDStoreOptions
	ClusterName  string
	BasePath     string
	AbsolutePath string
}

func NewCRDStore(o *CRDStoreOptions) *CRDStore {
	return &CRDStore{
		options:      o,
		ClusterName:  o.ClusterName,
		BasePath:     o.BasePath,
		AbsolutePath: fmt.Sprintf("%s/%s", o.BasePath, o.ClusterName),
	}
}

type crdClientMeta struct {
	client    *client.ClusterAPIV1Alpha1Client
	clientset *apiextensionsclient.Clientset
}

func getClientMeta(cluster *cluster.Cluster) (*crdClientMeta, error) {
	kubeConfigPath := kubeconfig.GetKubeConfigPath(cluster)
	client, err := util.NewApiClient(kubeConfigPath)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	cs, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	clientMeta := &crdClientMeta{
		client:    client,
		clientset: cs,
	}
	return clientMeta, nil
}

func (crds *CRDStore) Commit(c *cluster.Cluster) error {
	cm, err := getClientMeta(c)
	if err != nil {
		return err
	}
	success := false
	for i := 0; i <= RetryAttempts; i++ {
		if _, err = v1alpha1.CreateClustersCRD(cm.clientset); err != nil {
			time.Sleep(SleepSecondsPerAttempt * time.Second)
			continue
		}
		success = true
		break
	}

	if !success && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("Error creating Clusters CRD: %v", err)
	}

	_, err = cm.client.Clusters().Create(c.ClusterAPI)
	if err != nil {
		return fmt.Errorf("Unable to record clusters: %v", err)
	}
	logger.Info("Declaring cluster: %v", c.Name)

	success = false
	for i := 0; i <= RetryAttempts; i++ {
		if _, err = v1alpha1.CreateMachinesCRD(cm.clientset); err != nil {
			time.Sleep(SleepSecondsPerAttempt * time.Second)
			continue
		}
		success = true
		break
	}
	if !success && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("Error creating Machines CRD: %v", err)
	}
	for _, ms := range c.MachineSets {
		if ms.Spec.Replicas == nil {
			continue
		}
		r := int(*ms.Spec.Replicas)
		for i := 0; i <= r; i++ {
			calculatedName := fmt.Sprintf("%s-%d", ms.Name, i)
			machine := &v1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: calculatedName,
				},
				Spec: v1alpha1.MachineSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name: calculatedName,
					},
					ProviderConfig: ms.Spec.Template.Spec.ProviderConfig,
				},
			}
			_, err = cm.client.Machines().Create(machine)
			if err != nil {
				return fmt.Errorf("Unable to record machine: %v", err)
			}
			logger.Info("Declaring machine: %s", calculatedName)
		}

	}

	//NewAPIClient
	//if c == nil {
	//	return fmt.Errorf("Nil cluster spec")
	//}
	//bytes, err := yaml.Marshal(c)
	//if err != nil {
	//	return err
	//}
	//crds.write(state.ClusterYamlFile, bytes)
	return nil
}

func (crds *CRDStore) Exists() bool {
	//if _, err := os.Stat(crds.AbsolutePath); os.IsNotExist(err) {
	//	return false
	//}
	return true
}

func (crds *CRDStore) write(relativePath string, data []byte) error {
	//fqn := fmt.Sprintf("%s/%s", crds.AbsolutePath, relativePath)
	//os.MkdirAll(path.Dir(fqn), 0700)
	//fo, err := os.Create(fqn)
	//if err != nil {
	//	return err
	//}
	//defer fo.Close()
	//_, err = io.Copy(fo, strings.NewReader(string(data)))
	//if err != nil {
	//	return err
	//}
	return nil
}

func (crds *CRDStore) Read(relativePath string) ([]byte, error) {
	//
	var bytes []byte
	return bytes, nil
}

func (crds *CRDStore) ReadStore() ([]byte, error) {
	//return crds.Read(state.ClusterYamlFile)
	var bytes []byte
	return bytes, nil
}

func (crds *CRDStore) Rename(existingRelativePath, newRelativePath string) error {
	//return os.Rename(existingRelativePath, newRelativePath)
	return nil
}

func (crds *CRDStore) Destroy() error {
	//logger.Warning("Removing path [%s]", crds.AbsolutePath)
	//return os.RemoveAll(crds.AbsolutePath)
	return nil
}

func (crds *CRDStore) GetCluster() (*cluster.Cluster, error) {
	//configBytes, err := crds.Read(state.ClusterYamlFile)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return crds.BytesToCluster(configBytes)
	return &cluster.Cluster{}, nil
}

func (crds *CRDStore) BytesToCluster(bytes []byte) (*cluster.Cluster, error) {
	cluster := &cluster.Cluster{}
	err := yaml.Unmarshal(bytes, cluster)
	if err != nil {
		return cluster, err
	}
	return cluster, nil
}

func (crds *CRDStore) List() ([]string, error) {

	var stateList []string
	//
	//files, err := ioutil.ReadDir(crds.options.BasePath)
	//if err != nil {
	//	return stateList, err
	//}
	//
	//for _, file := range files {
	//	stateList = append(stateList, file.Name())
	//}

	return stateList, nil
}
