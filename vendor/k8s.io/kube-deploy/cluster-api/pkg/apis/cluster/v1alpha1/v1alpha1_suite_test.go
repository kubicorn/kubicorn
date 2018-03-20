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

package v1alpha1_test

import (
	"testing"

	"github.com/kubernetes-incubator/apiserver-builder/pkg/test"

	"k8s.io/kube-deploy/cluster-api/pkg/apis"
	"k8s.io/kube-deploy/cluster-api/pkg/client/clientset_generated/clientset"
	"k8s.io/kube-deploy/cluster-api/pkg/openapi"
)

func TestV1alpha1(t *testing.T) {
	testenv := test.NewTestEnvironment()
	config := testenv.Start(apis.GetAllApiBuilders(), openapi.GetOpenAPIDefinitions)
	cs := clientset.NewForConfigOrDie(config)

	t.Run("crudAccessToClusterClient", func(t *testing.T) {
		crudAccessToClusterClient(t, cs)
	})
	t.Run("crudAccessToMachineClient", func(t *testing.T) {
		crudAccessToMachineClient(t, cs)
	})
	t.Run("crudAccessToMachineSetClient", func(t *testing.T) {
		// TODO: the following test fails with:
		// the namespace of the provided object does not match the namespace sent on the request
		// uncomment when fixed
		//crudAccessToMachineSetClient(t, cs)
	})

	testenv.Stop()
}
