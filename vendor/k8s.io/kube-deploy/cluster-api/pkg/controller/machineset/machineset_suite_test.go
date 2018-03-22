/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machineset_test

import (
	"testing"

	"github.com/kubernetes-incubator/apiserver-builder/pkg/test"

	"k8s.io/kube-deploy/cluster-api/pkg/apis"
	"k8s.io/kube-deploy/cluster-api/pkg/client/clientset_generated/clientset"
	cfg "k8s.io/kube-deploy/cluster-api/pkg/controller/config"
	"k8s.io/kube-deploy/cluster-api/pkg/controller/machineset"
	"k8s.io/kube-deploy/cluster-api/pkg/controller/sharedinformers"
	"k8s.io/kube-deploy/cluster-api/pkg/openapi"
)

func TestMachineSet(t *testing.T) {
	testenv := test.NewTestEnvironment()
	config := testenv.Start(apis.GetAllApiBuilders(), openapi.GetOpenAPIDefinitions)
	cs := clientset.NewForConfigOrDie(config)
	cfg.ControllerConfig.Cloud = "test"

	shutdown := make(chan struct{})
	si := sharedinformers.NewSharedInformers(config, shutdown)
	controller := machineset.NewMachineSetController(config, si)
	controller.Run(shutdown)

	t.Run("machineSetControllerReconcile", func(t *testing.T) {
		machineSetControllerReconcile(t, cs, controller)
	})

	close(shutdown)
	testenv.Stop()
}
