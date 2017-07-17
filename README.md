[![Go Report Card](https://goreportcard.com/badge/github.com/kris-nova/klone)](https://goreportcard.com/report/github.com/nivenly/kamp)  [![GoDoc Widget]][GoDoc]

[GoDoc]: https://godoc.org/k8s.io/kops
[GoDoc Widget]: https://godoc.org/k8s.io/kops?status.svg

# kubicorn

Create, manage, snapshot, and scale Kubernetes infrastructure in the public cloud.

**Phonetic pronunciation**: `KEW - BHIK - OH - AR - IN`

## About

`kubicorn` is an **unofficial** project that solves the Kubernetes infrastructure problem and gives users a rich golang library to work with infrastructure.

`kubicorn` is a project that helps a user manage cloud infrastructure for Kubernetes.
With `kubicorn` a user can create new clusters, modify and scale them, and take a snapshot of their cluster at any time.

**NOTE:** This is a work-in-progress, we do not consider it production ready.
Use at your own risk and if you're as excited about it as we are, maybe you want to join us on the #kubicorn channel in the Gophers Slack community.

## Installing

```
$ go get github.com/kris-nova/kubicorn
```
## How is Kubicorn different?

1) We use kubeadm to bootstrap our clusters
2) We strive for developer empathy, and clean and simple code
3) We strive for operational empathy, and clean and simple user experience
4) We start with struct literals for profiles, and then marshal into an object
5) We offer the tooling as a library, more than a command line tool
6) We are atomic, and will un-do any work if there is an error
7) We run on many operating systems
8) We allow users to define their own arbitrary bootstrap logic
9) We have no guarantee that anything works, ever, use at your own risk
10) We have no dependency on DNS
11) We believe in snapshots, and that a user should be able to capture a cluster, and move it

# Concepts

### Create

`kubicorn` let's a user create a Kubernetes cluster in a cloud of their choice.

### Apply

Define what you want, then apply it. That simple.

### Adopt

`kubicorn` can adopt any Kubernetes cluster at any time.

### Scale

`kubicorn` is powered by a state enforcement pattern.
A user defines the intended state of Kubernetes infrastructure, and `kubicorn` can enforce the intended state.

### Snapshot

`kubicorn` allows a user to take a snapshot of a Kubernetes cluster, and run the image in any cloud at any time.
A snapshot is compressed file that will represent intedend infrastructure **and** intended application definitions.
Take a snap, save a snap, deploy a snap.

### Enforce

`kubicorn` is built as a library and a framework. Thus allowing it to be easily vendored into operator and controller patterns to enforce indeded state of infrastructure.

# Documentation

### AWS

| Name                       | Description                                                 | Link                                                                   |
| ---------------------------| ----------------------------------------------------------- |:----------------------------------------------------------------------:|
| **Walkthrough**            | A walkthrough guide on install Kubernetes 1.7.0 in AWS      | [walkthrough](docs/aws/walkthrough.md)                                 |
| **Video**                  | A step by step video of using Kubicorn in AWS               | [video](https://www.useloom.com/share/a0afd5034e654b0b8d6785a5fa8ec754)|
