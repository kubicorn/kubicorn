---
layout: documentation
title: Setting up Kubernetes in Azure
date: 2017-10-01
doctype: general
---

## Introduction

This document explains how `kubicorn` project works, the project's structure and information you need to know so you can easily start contributing.

## Glossary

This chapter explains the most important concepts.

* [Cluster API](https://github.com/kris-nova/kubicorn/tree/master/apis) — universal (cloud provider agnostic) representation of a Kubernetes cluster. It defines every part of an cluster, including infrastructure parts such as virtual machines, networking units and firewalls. The matching of universal representation of the cluster to the representation for specific cloud provider is done as the part of Reconciling process.
* [State](https://github.com/kris-nova/kubicorn/blob/master/state/README.md) — representation of an specific Kubernetes cluster. It's used in reconciling process to create the cluster.
* [State store](https://github.com/kris-nova/kubicorn/blob/master/state/README.md) — place where state is stored. Currently, we only support states located on the disk and in [YAML](https://github.com/kris-nova/kubicorn/tree/master/state/fs) or [JSON](https://github.com/kris-nova/kubicorn/tree/master/state/jsonfs) format. We're looking forward to implementing Git and S3 state stores.
* [Reconciler](https://github.com/kris-nova/kubicorn/tree/master/cloud) — the core of the project and place where provisioning logic is located. It matches cloud provider agnstoic Cluster definition to the specific cloud provider definition, which is used to provision cluster. It takes care of provisioning new clusters, destorying the old ones and keeping the consistency between Actual and Expected states.
* Actual state — the representation of current resources in the cloud.
* Expected state — the representation of intended resources in the cloud.
* [Bootstrap scripts](https://github.com/kris-nova/kubicorn/tree/master/bootstrap) — Bootstrap scripts are provided as the `user data` on the cluster creation to install dependencies and create the cluster. They're provided as Bash scripts, so you can easily create them without Go knowledge. You can also inject values in the reconciling process, per your needs.
* [VPN Boostrap scripts](https://github.com/kris-nova/kubicorn/tree/master/bootstrap/vpn) — to improve security of our cluster, we create VPN server on master, and connect every node using it. Some cloud providers, such as DigitalOcean, doesn't provide real private networking between Droplets, so we want master and nodes can only communicate between themselves, and with the Internet only on selected ports.
* [Profile](https://github.com/kris-nova/kubicorn/blob/master/profiles/README.md) — profile is a unique representation of a cluster written in Go. Profiles containts the all information needed to create an cluster, such as: cluster name, cloud provider, VM size, SSH key, network and firewall configurations...

## Project structure

Project is contained from the several packages. 

The most important package is the [`cloud`](https://github.com/kris-nova/kubicorn/tree/master/cloud) which contains Reconciler interface and Reconciler implementations for each cloud provider. Currently, we have four cloud provider implementation: [Amazon](https://github.com/kris-nova/kubicorn/tree/master/cloud/amazon), [DigitalOcean](https://github.com/kris-nova/kubicorn/tree/master/cloud/digitalocean), [Google](https://github.com/kris-nova/kubicorn/tree/master/cloud/google) and [Azure (WIP)](https://github.com/kris-nova/kubicorn/pull/327).

The Cluster API is located in the [`apis`](https://github.com/kris-nova/kubicorn/tree/master/apis) package.

The Bootstrap Scripts are located in the [`bootstrap`](https://github.com/kris-nova/kubicorn/tree/master/bootstrap) directory of the project. It also contains [`vpn`](https://github.com/kris-nova/kubicorn/tree/master/bootstrap/vpn) sub-directory with VPN implementations.

Default profiles are located in the [`profiles`](https://github.com/kris-nova/kubicorn/tree/master/profiles) package. Currently, we have Ubuntu profiles available for Amazon, DigitalOcean and GCE, and CentOS profiles available for Amazon and DigitalOcean.

State store definitions are located in the [`state`](https://github.com/kris-nova/kubicorn/tree/master/state) package.

We have two type of tests — CI tests and E2E tests. CI tests are regular Go tests, while E2E tests are run against real cloud infrastucture and it can cost money. E2E tests are available in the [`test`](https://github.com/kris-nova/kubicorn/tree/master/test) package.

The [`cutil`](https://github.com/kris-nova/kubicorn/tree/master/cutil) directory contains many useful, helper packages which are used to do various tasks, such as: copy the file from the VM, create `kubeadm` token, logger implementation...

The [`cmd`](https://github.com/kris-nova/kubicorn/tree/master/cmd) package is the CLI implementation for `kubicorn`. We use [`cobra`](https://github.com/spf13/cobra) package to create the CLI.

## Reconciler

TODO

## Where should I start?

First, you should take a look at the [examples](https://github.com/kris-nova/kubicorn/tree/master/examples). It should explain you which steps are taken to create a Kubernetes cluster.

Once you get familiar with the process, we recommend taking a look at the specific implementation of an Reconciler.

## Why where are doing stuff this way?

In this document, we'll not explain reasoning and decisions involved in this project. If you are interested in the details, you should take a look at the [Cloud Native Infrastructure](http://shop.oreilly.com/product/0636920075837.do) book by Justin Garrison and Kris Nova.
Also, if you have any question, feel free to create an issue or ask us on the `kubicorn` channel at the [Gophers Slack](http://kubicorn.io/).