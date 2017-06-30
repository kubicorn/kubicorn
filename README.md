[![Go Report Card](https://goreportcard.com/badge/github.com/kris-nova/klone)](https://goreportcard.com/report/github.com/nivenly/kamp)

# kubicorn

Create, Manage, Image, and Scale Kubernetes infrastructure in the cloud.

## About

`kubicorn` is a project that solves the Kubernetes infrastructure problem.

`kubicorn` is a command line tool that helps a user manage cloud infrastructure for Kubernetes.
With `kubicorn` a user can create new clusters, modify and scale them, and take a snapshot of their cluster at any time.


### Create

`kubicorn` let's a user create a Kubernetes cluster in a cloud of their choice.

### Adopt

`kubicorn` can adopt any Kubernetes cluster at any time.

### Scale

`kubicorn` is powered by a state enforcement model.
A user defines the intended state of Kubernetes infrastructure, and `kubicorn` can enforce the intended state.

### Snapshot

`kubicorn` allows a user to take a snapshot of a Kubernetes cluster, and run the image in any cloud at any time.


# Supported Clouds

<p align="left">
  <img src="docs/img/aws.png" width="200"> </image>
</p>

 - Highly Available (HA)
 - Public topology
 - Private topology

<p align="left">
  <img src="docs/img/azure.png" width="200"> </image>
</p>

 - Public topology
 - Private topology

<p align="left">
  <img src="docs/img/google.png" width="200"> </image>
</p>

 - Public topology
 - Private topology


