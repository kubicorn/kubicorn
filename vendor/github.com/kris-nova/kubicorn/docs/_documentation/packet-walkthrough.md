---
layout: documentation
title: Setting up Kubernetes in Packet
date: 2017-11-25
doctype: packet
---

In this document, we're going to show you how to use `kubicorn` to ramp up a Kubernetes cluster in [Packet](https://www.packet.net), use it and tear it down again.

As a prerequisite, you need to have `kubicorn` installed. Since we don't have binary releases yet, we assume you've got Go installed and simply do:

#### Installing

```
$ go get github.com/kubicorn/kubicorn
```

The first thing you will do now is to define the cluster resources.
For this, you need to select a certain profile. Of course, once you're more familiar with `kubicorn`, you can go ahead and extend existing profiles or create new ones.
In the following we'll be using an existing profile called `packet`, which is a profile for a cluster in Packet based on Ubuntu 16.04 LTS servers.

#### Creating

Now execute the following command:

```
$ kubicorn create myfirstk8s --profile packet
```

Verify that `kubicorn create` did a good job by executing:

```
$ cat _state/myfirstk8s/cluster.yaml
```

Feel free to tweak the configuration to your liking here.

#### Authenticating

We're now in a position to have the cluster resources defined, locally, based on the selected profile.
Next we will apply the so defined resources using the `apply` command, but before we do that we'll set up the access to Packet.
You will need a Packet API token.
You can create an API token in the [Packet portal](https://app.packet.net/portal#/api-keys).

Next, export the environment variable `PACKET_APITOKEN` so that `kubicorn` can pick it up in the next step:

```
$ export PACKET_APITOKEN=*****************************************
```

Also, make sure that the public SSH key for your Packet account is called `id_rsa.pub`, which is the default in the above profile:

```
$ ls -al ~/.ssh/id_rsa.pub
-rw-------@ 1 mhausenblas  staff   754B 20 Mar 04:03 /Users/mhausenblas/.ssh/id_rsa.pub
```

#### Applying

With the access set up, we can now apply the resources we defined in the first step.
This actually creates resources in Packet. Up to now we've only been working locally.

So, execute:

```
$ kubicorn apply myfirstk8s
```

Now `kubicorn` will reconcile your intended state against the actual state in the cloud, thus creating a Kubernetes cluster.
A `kubectl` configuration file (kubeconfig) will be created or appended for the cluster on your local filesystem.
You can now `kubectl get nodes` and verify that Kubernetes 1.7.0 is now running.
You can also `ssh` into your instances using the example command found in the output from `kubicorn`

#### Deleting

To delete your cluster run:

```
$ kubicorn delete myfirstk8s
```

Congratulations, you're an official `kubicorn` user now and might want to dive deeper,
for example, learning how to define your own [profiles](https://github.com/kubicorn/kubicorn/tree/master/profiles).
