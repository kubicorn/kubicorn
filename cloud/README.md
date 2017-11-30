# Cloud package

The `cloud` package contains the Reconciler interface and Reconciler implementations for each cloud provider.

## Overview
The tasks covered in this package include

* *rendering* a universal, cloud-provider agnostic cluster representation into a representation for a specific cloud provider
* obtaining the *expected* state of the cluster
* obtaining the *actual* state of the cluster
* comparing *expected* and *actual* and *applying* changes

The *Apply* function does most of the work, including

* Building Bootstrap scripts and injecting values at the runtime
* Creating the appropriate resources for Master node
* Obtaining needed information for Node creation
* Creating Nodes
* Downloading the .kubeconfig file for the cluster from the master node

See the [Kubicorn project walkthrough](http://kubicorn.io/documentation/readme.html) for a more detailed account of these components and how they interrelate.

## Adding a Cloud Provider
To add a new cloud provider, please see the document [Adding a new cloud provider](../docs/_documentation/cloud-providers.md).
