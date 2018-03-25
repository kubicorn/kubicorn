---
layout: documentation
title: Cloud Providers
date: 2017-11-19
doctype: general
---


## Adding a new cloud provider
In order to add a new cloud provider, you need to implement several key functions and interfaces:

0. Create a name for this cloud provider as a string, e.g. `"mycloud"` and keep it handy.
1. Create a `const` representing the name of the cloud in [cluster.go](https://github.com/kubicorn/kubicorn/blob/master/apis/cluster/cluster.go#L21), with the `const` name being `Cloud<name>` and the value being the name from the first step.
2. Create a subdirectory in this [cloud](https://github.com/kubicorn/kubicorn/tree/master/cloud) directory named for the provider providing a cloud implementation. The directory name should be the name of the provider from the first step. A cloud implementation is all of the constants, interfaces and functions necessary to execute commands against a cloud provider.
3. Create a subdirectory in the [profiles directory](https://github.com/kubicorn/kubicorn/tree/master/profiles) named for the provider. The directory name should be the same as the name of the provider from the first step.
4. Create one or more profiles for the provider in the [profiles directory](https://github.com/kubicorn/kubicorn/tree/master/profiles) for the provider. A profile contains all of the necessary information to create a specific implementation in a given cloud provider, e.g. instance size, ssh key, etc.
5. Add the implemented profiles as profile options to `profileMapIndexed` in [create.go](https://github.com/kubicorn/kubicorn/blob/master/cmd/create.go). Be sure to `import` the necessary package from `profiles/`
6. Add the provider as a `known.Cloud` to [reconciler.go](https://github.com/kubicorn/kubicorn/blob/master/pkg/reconciler.go)
7. Add any bootstrap scripts required in [bootstrap/](https://github.com/kubicorn/kubicorn/tree/master/bootstrap/). As a general rule, the name of the script should be `<provider>_k8s_<profile>_master.sh` or `<provider>_k8s_<profile>_node.sh`

### Profile
A profile is expected to return a single function, the "profile function". You can name that function anything you want. In `create.go`, you will save that function as `profileFunc` and then call it.

For example, the following code is how the AWS Ubuntu profile function is saved:

```go
import (
	"github.com/kubicorn/kubicorn/profiles/amazon"
)
var profileMapIndexed = map[string]profileMap{
	// other stuff skipped here
	"amazon": {
		profileFunc: amazon.NewUbuntuCluster,
		description: "Ubuntu on Amazon",
	},
	// lots of other stuff skipped here
}
```

This means that when someone runs `kubicorn create mycluster --profile amazon`, it will use the profile function `NewUbuntuCluster` from the package `github.com/kubicorn/kubicorn/profiles/amazon`.

Notice that it doesn't matter _what_ the function was named. It could have been named `TheUglyDuckling`, as it is saved as `profileFunc` in the `profileMapIndexed` map, under the key `"amazon"`.

If you have multiple profiles - perhaps different VM sizes or different OSes - you simply use a different profile function and save it under a different key un the `profileMapIndexed` map.

The profile function itself is expected to return `*cluster.Cluster` from the package [apis/cluster](https://github.com/kubicorn/kubicorn/tree/master/apis/cluster). This is a struct with all of the information necessary to create the resources for the cluster.

### Cloud Provider Implementation
The cloud provider profile is the mapping between the generic cloud API used to perform functions on cloud providers, and the specific API calls required for your cloud provider.

The cloud provider implementation is expect to provide a function that returns an implementation of [Model](https://github.com/kubicorn/kubicorn/blob/master/cloud/interface.go#L39). This typically is done using a function with the following signature:

```go
GetModel(known *cluster.Cluster) cloud.Model
```

The function should use the `known` parameter to provide profile-specific information.

You can name the function anything you want.

This model factory should be called from within the appropriate provider-specific block in [reconciler.go](https://github.com/kubicorn/kubicorn/blob/master/pkg/reconciler.go). the block is expected to do the following:

1. Initialize any APIs or SDKs necessary.
2. Call `cloud.NewAtomicReconciler(known, ModelFunc(known))` , where `ModelFunc` is the model function for your cloud provider implementation.
3. Return the results of the above call

Once called, the `Reconciler` has a `Model` that is specific to your cloud provider and has been initialized with appropriate profile information. It then will use the methods on that interface to make calls to your cloud provider.

The `Reconciler` then will call `model.Resources()` to get an array of all known `Resource` implementations for the cloud provider, on which it can call `Actual`, `Expected`, `Apply` and `Delete`.
