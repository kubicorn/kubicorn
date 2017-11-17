/*
Copyright 2016 The Kubernetes Authors.

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

package discovery

import (
<<<<<<< HEAD
	"reflect"
=======
	"fmt"
>>>>>>> Initial dep workover

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

<<<<<<< HEAD
// UnstructuredObjectTyper provides a runtime.ObjectTyper implementation for
// runtime.Unstructured object based on discovery information.
type UnstructuredObjectTyper struct {
	registered map[schema.GroupVersionKind]bool
	typers     []runtime.ObjectTyper
}

// NewUnstructuredObjectTyper returns a runtime.ObjectTyper for
// unstructured objects based on discovery information. It accepts a list of fallback typers
// for handling objects that are not runtime.Unstructured. It does not delegate the Recognizes
// check, only ObjectKinds.
func NewUnstructuredObjectTyper(groupResources []*APIGroupResources, typers ...runtime.ObjectTyper) *UnstructuredObjectTyper {
	dot := &UnstructuredObjectTyper{
		registered: make(map[schema.GroupVersionKind]bool),
		typers:     typers,
	}
=======
// UnstructuredObjectTyper provides a runtime.ObjectTyper implmentation for
// runtime.Unstructured object based on discovery information.
type UnstructuredObjectTyper struct {
	registered map[schema.GroupVersionKind]bool
}

// NewUnstructuredObjectTyper returns a runtime.ObjectTyper for
// unstructred objects based on discovery information.
func NewUnstructuredObjectTyper(groupResources []*APIGroupResources) *UnstructuredObjectTyper {
	dot := &UnstructuredObjectTyper{registered: make(map[schema.GroupVersionKind]bool)}
>>>>>>> Initial dep workover
	for _, group := range groupResources {
		for _, discoveryVersion := range group.Group.Versions {
			resources, ok := group.VersionedResources[discoveryVersion.Version]
			if !ok {
				continue
			}

			gv := schema.GroupVersion{Group: group.Group.Name, Version: discoveryVersion.Version}
			for _, resource := range resources {
				dot.registered[gv.WithKind(resource.Kind)] = true
			}
		}
	}
	return dot
}

<<<<<<< HEAD
=======
// ObjectKind returns the group,version,kind of the provided object, or an error
// if the object in not runtime.Unstructured or has no group,version,kind
// information.
func (d *UnstructuredObjectTyper) ObjectKind(obj runtime.Object) (schema.GroupVersionKind, error) {
	if _, ok := obj.(runtime.Unstructured); !ok {
		return schema.GroupVersionKind{}, fmt.Errorf("type %T is invalid for dynamic object typer", obj)
	}

	return obj.GetObjectKind().GroupVersionKind(), nil
}

>>>>>>> Initial dep workover
// ObjectKinds returns a slice of one element with the group,version,kind of the
// provided object, or an error if the object is not runtime.Unstructured or
// has no group,version,kind information. unversionedType will always be false
// because runtime.Unstructured object should always have group,version,kind
// information set.
func (d *UnstructuredObjectTyper) ObjectKinds(obj runtime.Object) (gvks []schema.GroupVersionKind, unversionedType bool, err error) {
<<<<<<< HEAD
	if _, ok := obj.(runtime.Unstructured); ok {
		gvk := obj.GetObjectKind().GroupVersionKind()
		if len(gvk.Kind) == 0 {
			return nil, false, runtime.NewMissingKindErr("object has no kind field ")
		}
		if len(gvk.Version) == 0 {
			return nil, false, runtime.NewMissingVersionErr("object has no apiVersion field")
		}
		return []schema.GroupVersionKind{gvk}, false, nil
	}
	var lastErr error
	for _, typer := range d.typers {
		gvks, unversioned, err := typer.ObjectKinds(obj)
		if err != nil {
			lastErr = err
			continue
		}
		return gvks, unversioned, nil
	}
	if lastErr == nil {
		lastErr = runtime.NewNotRegisteredErrForType(reflect.TypeOf(obj))
	}
	return nil, false, lastErr
=======
	gvk, err := d.ObjectKind(obj)
	if err != nil {
		return nil, false, err
	}

	return []schema.GroupVersionKind{gvk}, false, nil
>>>>>>> Initial dep workover
}

// Recognizes returns true if the provided group,version,kind was in the
// discovery information.
func (d *UnstructuredObjectTyper) Recognizes(gvk schema.GroupVersionKind) bool {
	return d.registered[gvk]
}

<<<<<<< HEAD
=======
// IsUnversioned returns false always because runtime.Unstructured objects
// should always have group,version,kind information set. ok will be true if the
// object's group,version,kind is api.Registry.
func (d *UnstructuredObjectTyper) IsUnversioned(obj runtime.Object) (unversioned bool, ok bool) {
	gvk, err := d.ObjectKind(obj)
	if err != nil {
		return false, false
	}

	return false, d.registered[gvk]
}

>>>>>>> Initial dep workover
var _ runtime.ObjectTyper = &UnstructuredObjectTyper{}
