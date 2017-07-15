# Walkthrough

In the following, we gonna show you how to use `kubicorn` to ramp up a Kubernetes cluster in AWS, use it and tear it down again.

As a prerequisite, you need to have `kubicorn` installed. Since we don't have binary releases yet, we assume you've got Go installed and simply do:

```
$ go get github.com/kris-nova/kubicorn
```

The first thing you will do now is to define the cluster resources. For this, you need to select a certain profile. Of course, once you're more familiar with `kubicorn`, you can go ahead and extend existing profiles or create new ones.
In the following we'll be using an existing profile called `aws`, which is—surprise, surprise—a profile for a cluster in AWS.

Now execute the following command:

```
$ kubicorn create --name myfirstk8s --profile aws
```

Note that `kubicorn` executes silently as long as there are no errors but you can always use the `--verbose` flag to increase the information output, for example `kubicorn create --verbose 4` gives you detailed info on what it is doing at any step.

Verify that `kubicorn create` did a good job by executing:

```
$ cat _state/myfirstk8s/cluster.yaml
```

We're now in a position to have the cluster resources defined, locally, based on the selected profile. Next we will apply the so defined resources using the `apply` command, but before we do that we'll set up the access to AWS. You might want to create a new [IAM user](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html) for this with the following permissions:

![AWS IAM permissions required for kubicorn](img/aws-iam-user-perm-screen-shot.png)

Next, export the two environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` so that `kubicorn` can pick it up in the next step:

```
$ export AWS_ACCESS_KEY_ID=***************
$ export AWS_SECRET_ACCESS_KEY=*****************************************
```

Also, make sure that the public SSH key for your AWS account is called `id_rsa.pub`, which is the default in above profile:

```
$ ls -al ~/.ssh/id_rsa.pub
-rw-------@ 1 mhausenblas  staff   754B 20 Mar 04:03 /Users/mhausenblas/.ssh/id_rsa.pub
```

With the access set up, we can now apply the resources we defined in the first step. This actually creates resources in AWS. Up to now we've only been working locally.

So, execute:

```
$ kubicorn apply --name myfirstk8s
```

This command again will execute silently. In order to verify if it has been working correctly, go to your AWS console,
select the `us-west-2` region (default in the profile we used) and you should see something like the following:

![AWS EC2 post-creation](img/aws-ec2-post-create-screen-shot.png)

Wait a few minutes until the dust has settled, note the IP of the node (in our case) `myfirstk8s.amazon-master`
and use this to ssh into the master node where the Kubernetes control plane lives:

```
$ ssh -i ~/.ssh/id_rsa ubuntu@52.39.36.164
ubuntu@ip-10-0-0-216:~$ sudo docker ps
CONTAINER ID        IMAGE                                                                                                                            COMMAND                  CREATED             STATUS              PORTS               NAMES
0b4e60d9b645        gcr.io/google_containers/kube-proxy-amd64@sha256:ce0bd283fbc5c217d3a81c5917996e6dddfe7110437a712f84a94d2d5912214d                "/usr/local/bin/kube-"   28 minutes ago      Up 28 minutes                           k8s_kube-proxy_kube-proxy-qbn2g_kube-system_204158a9-698d-11e7-8d2a-0655e7e5e6f2_0
b3a2dbd5f646        gcr.io/google_containers/pause-amd64:3.0                                                                                         "/pause"                 28 minutes ago      Up 28 minutes                           k8s_POD_kube-proxy-qbn2g_kube-system_204158a9-698d-11e7-8d2a-0655e7e5e6f2_0
8185fb868fce        gcr.io/google_containers/kube-controller-manager-amd64@sha256:c119c0647ad980627b1d57c7e9a7fa0fc4af09345b8f46b64be56e36d97b1402   "kube-controller-mana"   29 minutes ago      Up 29 minutes                           k8s_kube-controller-manager_kube-controller-manager-ip-10-0-0-216_kube-system_6fefbb5265a660beb7e6c1df8c50fb8d_0
5b2a55d9574b        gcr.io/google_containers/etcd-amd64@sha256:d83d3545e06fb035db8512e33bd44afb55dea007a3abd7b17742d3ac6d235940                      "etcd --data-dir=/var"   29 minutes ago      Up 29 minutes                           k8s_etcd_etcd-ip-10-0-0-216_kube-system_fcc8d05cc29da0dc1f9e34c2e2cefa6e_0
ab37136aa4ce        gcr.io/google_containers/kube-apiserver-amd64@sha256:73d4a6883a4f4f78fce1829afad2dce54c31af2ec689d45728df11506283f1a2            "kube-apiserver --all"   29 minutes ago      Up 29 minutes                           k8s_kube-apiserver_kube-apiserver-ip-10-0-0-216_kube-system_8e22049ceee3692cd7479170554c76c5_0
6ba4dcba12d2        gcr.io/google_containers/kube-scheduler-amd64@sha256:85386a8929ac79e944bbd087444b1f3b07d9c5dc1d0c1b3d6ede3effa611378d            "kube-scheduler --lea"   29 minutes ago      Up 29 minutes                           k8s_kube-scheduler_kube-scheduler-ip-10-0-0-216_kube-system_b08d9a4015830eea5e3141df2788b07f_0
c3de45792578        gcr.io/google_containers/pause-amd64:3.0                                                                                         "/pause"                 29 minutes ago      Up 29 minutes                           k8s_POD_kube-controller-manager-ip-10-0-0-216_kube-system_6fefbb5265a660beb7e6c1df8c50fb8d_0
c83ddac9b182        gcr.io/google_containers/pause-amd64:3.0                                                                                         "/pause"                 29 minutes ago      Up 29 minutes                           k8s_POD_etcd-ip-10-0-0-216_kube-system_fcc8d05cc29da0dc1f9e34c2e2cefa6e_0
c270dc72eb20        gcr.io/google_containers/pause-amd64:3.0                                                                                         "/pause"                 29 minutes ago      Up 29 minutes                           k8s_POD_kube-apiserver-ip-10-0-0-216_kube-system_8e22049ceee3692cd7479170554c76c5_0
cd6d9bba32be        gcr.io/google_containers/pause-amd64:3.0                                                                                         "/pause"                 29 minutes ago      Up 29 minutes                           k8s_POD_kube-scheduler-ip-10-0-0-216_kube-system_b08d9a4015830eea5e3141df2788b07f_0
```

With above command we've verified that indeed `kubicorn` has launched the control plane and we're good to go.

You can now use your freshly baked Kubernetes cluster and once you're done, you can free the resources using following command:

```
$ kubicorn delete --name myfirstk8s
```

Congratulations, you're an official `kubicorn` user now and might want to dive deeper,
for example, learning how to define your own [profiles](https://github.com/kris-nova/kubicorn/tree/master/profiles).
