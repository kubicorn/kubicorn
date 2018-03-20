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
package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "k8s.io/kube-deploy/cluster-api/pkg/apis/cluster/v1alpha1"
	scheme "k8s.io/kube-deploy/cluster-api/pkg/client/clientset_generated/clientset/scheme"
)

// MachineSetsGetter has a method to return a MachineSetInterface.
// A group's client should implement this interface.
type MachineSetsGetter interface {
	MachineSets(namespace string) MachineSetInterface
}

// MachineSetInterface has methods to work with MachineSet resources.
type MachineSetInterface interface {
	Create(*v1alpha1.MachineSet) (*v1alpha1.MachineSet, error)
	Update(*v1alpha1.MachineSet) (*v1alpha1.MachineSet, error)
	UpdateStatus(*v1alpha1.MachineSet) (*v1alpha1.MachineSet, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.MachineSet, error)
	List(opts v1.ListOptions) (*v1alpha1.MachineSetList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.MachineSet, err error)
	MachineSetExpansion
}

// machineSets implements MachineSetInterface
type machineSets struct {
	client rest.Interface
	ns     string
}

// newMachineSets returns a MachineSets
func newMachineSets(c *ClusterV1alpha1Client, namespace string) *machineSets {
	return &machineSets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the machineSet, and returns the corresponding machineSet object, and an error if there is any.
func (c *machineSets) Get(name string, options v1.GetOptions) (result *v1alpha1.MachineSet, err error) {
	result = &v1alpha1.MachineSet{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("machinesets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of MachineSets that match those selectors.
func (c *machineSets) List(opts v1.ListOptions) (result *v1alpha1.MachineSetList, err error) {
	result = &v1alpha1.MachineSetList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("machinesets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested machineSets.
func (c *machineSets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("machinesets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a machineSet and creates it.  Returns the server's representation of the machineSet, and an error, if there is any.
func (c *machineSets) Create(machineSet *v1alpha1.MachineSet) (result *v1alpha1.MachineSet, err error) {
	result = &v1alpha1.MachineSet{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("machinesets").
		Body(machineSet).
		Do().
		Into(result)
	return
}

// Update takes the representation of a machineSet and updates it. Returns the server's representation of the machineSet, and an error, if there is any.
func (c *machineSets) Update(machineSet *v1alpha1.MachineSet) (result *v1alpha1.MachineSet, err error) {
	result = &v1alpha1.MachineSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("machinesets").
		Name(machineSet.Name).
		Body(machineSet).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *machineSets) UpdateStatus(machineSet *v1alpha1.MachineSet) (result *v1alpha1.MachineSet, err error) {
	result = &v1alpha1.MachineSet{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("machinesets").
		Name(machineSet.Name).
		SubResource("status").
		Body(machineSet).
		Do().
		Into(result)
	return
}

// Delete takes name of the machineSet and deletes it. Returns an error if one occurs.
func (c *machineSets) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("machinesets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *machineSets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("machinesets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched machineSet.
func (c *machineSets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.MachineSet, err error) {
	result = &v1alpha1.MachineSet{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("machinesets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
