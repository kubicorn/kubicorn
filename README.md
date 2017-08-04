[![Build Status](https://travis-ci.org/kris-nova/kubicorn.svg?branch=master)](https://travis-ci.org/kris-nova/kubicorn) [![Go Report Card](https://goreportcard.com/badge/github.com/kris-nova/kubicorn)](https://goreportcard.com/report/github.com/kris-nova/kubicorn)
# kubicorn


<img src="docs/img/kubicorn-trans.png" width="360"> Create, manage, snapshot, and scale Kubernetes infrastructure in the public cloud. The `kubicorn` project is a library that is specifically designed to be vendored into a controller or operator for rock solid Kubernetes infrastructure management from Kubernetes itself!

**Phonetic pronunciation**: `KEW - BHIK - OH - AR - IN`

## About

`kubicorn` is an **unofficial** project that solves the Kubernetes infrastructure problem and gives users a rich golang library to work with infrastructure.

`kubicorn` is a project that helps a user manage cloud infrastructure for Kubernetes.
With `kubicorn` a user can create new clusters, modify and scale them, and take a snapshot of their cluster at any time.

**NOTE:** This is a work-in-progress, we do not consider it production ready.
Use at your own risk and if you're as excited about it as we are, maybe you want to join us on the #kubicorn channel in the Gophers Slack community.


<img src="https://github.com/ashleymcnamara/gophers/blob/master/NERDY.png" width="60"> Proudly packaged with Golang [dep](https://github.com/golang/dep)


# Installing

```bash
$ go get github.com/kris-nova/kubicorn
``` 

..or read the [Install Guide](docs/INSTALL.md).

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

`kubicorn` lets a user create a Kubernetes cluster in a cloud of their choice.

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

| Name                          | Description                                                 | Link                                                                            |
| ----------------------------- | ----------------------------------------------------------- |:-------------------------------------------------------------------------------:|
| **Install**                   | Install guide for Kubicorn CLI                              | [install](docs/INSTALL.md)                                                      |
| **Environmental Variables**   | Master list of supported environmental variables            | [envvars](docs/envar.md)                                                        |
| **Kops vs Kubicorn**          | Blog about kubicorn with comparison table                   | [nivenly.com/kubicorn](https://nivenly.com/kubicorn)                            |
| **AWS Walkthrough**           | A walkthrough guide on install Kubernetes 1.7 in AWS      | [walkthrough](docs/aws/walkthrough.md)                                          |
| **Digital Ocean Walkthrough** | A walkthrough guide on install Kubernetes 1.7 in D.O.     | [walkthrough](docs/do/walkthrough.md)                                           |
| **AWS Video**                 | A step by step video of using Kubicorn in AWS               | [video](https://www.useloom.com/share/a0afd5034e654b0b8d6785a5fa8ec754)         |
| **Tech N Talk Deep Dive**     | A technical deep dive courtesy of RedHat                    | [youtube](https://youtu.be/2DmUG0RgS70?list=PLaR6Rq6Z4IqfwXtKT7KeARRvxdvyLqG72) |
