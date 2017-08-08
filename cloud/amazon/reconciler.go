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

package amazon

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kris-nova/kubicorn/apis/cluster"
	"github.com/kris-nova/kubicorn/cloud"
	"github.com/kris-nova/kubicorn/cloud/amazon/awsSdkGo"
	"github.com/kris-nova/kubicorn/cloud/amazon/resources"
	"github.com/kris-nova/kubicorn/cutil/hang"
	"github.com/kris-nova/kubicorn/cutil/logger"
)

var sigCaught = false

type Reconciler struct {
	Known *cluster.Cluster
}

func NewReconciler(expected *cluster.Cluster) cloud.Reconciler {
	return &Reconciler{
		Known: expected,
	}
}

var model map[int]cloud.Resource

func (r *Reconciler) Init() error {
	sdk, err := awsSdkGo.NewSdk(r.Known.Location)
	if err != nil {
		return err
	}
	resources.Sdk = sdk
	model = ClusterModel(r.Known)
	return nil
}

func (r *Reconciler) GetActual() (*cluster.Cluster, error) {
	actualCluster := newClusterDefaults(r.Known)
	for i := 0; i < len(model); i++ {
		resource := model[i]
		actualResource, err := resource.Actual(r.Known)
		if err != nil {
			return nil, err
		}
		actualCluster, err = resource.Render(actualResource, actualCluster)
		if err != nil {
			return nil, err
		}
	}
	return actualCluster, nil
}

func (r *Reconciler) GetExpected() (*cluster.Cluster, error) {
	expectedCluster := newClusterDefaults(r.Known)
	for i := 0; i < len(model); i++ {
		resource := model[i]
		expectedResource, err := resource.Expected(r.Known)
		if err != nil {
			return nil, err
		}
		expectedCluster, err = resource.Render(expectedResource, expectedCluster)
		if err != nil {
			return nil, err
		}
	}
	return expectedCluster, nil
}

func cleanUp(cluster *cluster.Cluster, i int) error {
	logger.Warning("--------------------------------------")
	logger.Warning("Attempting to delete created resources!")
	logger.Warning("--------------------------------------")
	for j := i - 1; j >= 0; j-- {
		resource := model[j]
		createdResource := createdResources[j]
		err := resource.Delete(createdResource, cluster)
		if err != nil {
			j, err = destroyI(err, j)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

var createdResources = make(map[int]cloud.Resource)

func (r *Reconciler) Reconcile(actualCluster, expectedCluster *cluster.Cluster) (*cluster.Cluster, error) {
	newCluster := newClusterDefaults(r.Known)

	for i := 0; i < len(model); i++ {
		if sigCaught {
			cleanUp(newCluster, i)
			os.Exit(1)
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		go handleCtrlC(c)

		resource := model[i]
		expectedResource, err := resource.Expected(expectedCluster)
		if err != nil {
			return nil, err
		}
		actualResource, err := resource.Actual(actualCluster)
		if err != nil {
			return nil, err
		}
		appliedResource, err := resource.Apply(actualResource, expectedResource, newCluster)
		if err != nil {
			logger.Critical("Error during apply! Attempting cleaning: %v", err)
			err = cleanUp(newCluster, i)
			if err != nil {
				logger.Critical("Failure during cleanup! Abandoned resources!")
				return nil, err
			}
			return nil, nil
		}
		newCluster, err = resource.Render(appliedResource, newCluster)
		if err != nil {
			return nil, err
		}
		createdResources[i] = appliedResource
	}

	return newCluster, nil
}

var destroyRetryStrings = []string{
	"DependencyViolation:",
	"does not exist in default VPC",
}

var hg = &hang.Hanger{
	Ratio: 1,
}

func destroyI(err error, i int) (int, error) {
	hg.Hang()
	for _, retryString := range destroyRetryStrings {
		if strings.Contains(err.Error(), retryString) {
			logger.Debug("Retry failed delete: %v", err)
			time.Sleep(1 * time.Second)
			return i + 1, nil
		}
	}
	return 0, err
}

func (r *Reconciler) Destroy() error {
	for i := len(model) - 1; i >= 0; i-- {
		resource := model[i]
		actualResource, err := resource.Actual(r.Known)
		if err != nil {
			i, err = destroyI(err, i)
			if err != nil {
				return err
			}
			continue
		}
		err = resource.Delete(actualResource, r.Known)
		if err != nil {
			i, err = destroyI(err, i)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func newClusterDefaults(base *cluster.Cluster) *cluster.Cluster {
	new := &cluster.Cluster{
		Name:          base.Name,
		Cloud:         base.Cloud,
		Location:      base.Location,
		Network:       &cluster.Network{},
		SSH:           &cluster.SSH{},
		Values:        base.Values,
		KubernetesAPI: base.KubernetesAPI,
	}
	return new
}

func handleCtrlC(c chan os.Signal) {
	sig := <-c
	if sig == syscall.SIGINT {
		sigCaught = true
		logger.Warning("SIGINT! Why did you do that? Trying to rewind to clean up orphaned resources!")
	}
}
