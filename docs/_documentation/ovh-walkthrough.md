---
layout: documentation
title: Setting up Kubernetes in OVH
date: 2018-02-23
doctype: ovh
---

In the following, we're going to show you how to use `kubicorn` to ramp up a Kubernetes cluster in OVH Public Cloud, use it and tear it down again. OVH Public Cloud is based on Openstack.

The cluster will be running in a private network, using OVH vRack technology. Prior to anything else, you must make sure a vRack is associated to your cloud project.

Then you need to have `kubicorn` installed. Since we don't have binary releases yet, we assume you've got Go installed and simply do:

#### Installing

```bash
$ go get github.com/kubicorn/kubicorn
```

The first thing you will do now is to define the cluster resources.
For this, you need to select a certain profile. Of course, once you're more familiar with `kubicorn`, you can go ahead and extend existing profiles or create new ones.
In the following we'll be using an existing profile called `ovh`, which as it sounds, is a profile for a cluster in OVH.

#### Creating

Now execute the following command:

```bash
$ kubicorn create myfirstk8s --profile ovh
```

Verify that `kubicorn create` did a good job by executing:

```bash
$ cat _state/myfirstk8s/cluster.yaml
```

Feel free to tweak the configuration to your liking here.

#### Authenticating

We're now in a position to have the cluster resources defined, locally, based on the selected profile.
Next we will apply the so defined resources using the `apply` command, but before we do that we'll need our Openstack credentials exported in the environment.

You will need to retrieve the OpenRC file for the Openstack user you want to use from the customer interface, then run:

```bash
source openrc.sh
```

Also, make sure that the public SSH key in the above profile is correct, the default being `~/.ssh/id_rsa.pub`.

#### Applying

With the access set up, we can now apply the resources we defined in the first step.
This actually creates resources in OVH Public Cloud. Up to now we've only been working locally.

So, execute:

```bash
$ kubicorn apply myfirstk8s
```

Now `kubicorn` will reconcile your intended state against the actual state in the cloud, thus creating a Kubernetes cluster.
A `kubectl` configuration file (kubeconfig) will be created or appended for the cluster on your local filesystem.
You can now `kubectl get nodes` and verify that Kubernetes is now running.
You can also `ssh` into your instances using the example command found in the output from `kubicorn`

#### Deleting

To delete your cluster run:

```bash
$ kubicorn delete myfirstk8s
```

Congratulations, you're an official `kubicorn` user now and might want to dive deeper,
for example, learning how to define your own [profiles](https://github.com/kubicorn/kubicorn/tree/master/profiles).
