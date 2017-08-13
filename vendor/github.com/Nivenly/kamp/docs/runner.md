# Kubernetes Runner

### 1) Create Pod

The pod will have 2 containers.

 1. The kamp SSH server
 2. The arbitrary user container

### 2) Create Service

The service will expose an arbitrary node port that will map to the SSH server.
We will call this the `$NodePort`.
This will be the TCP port used to connect to the SSH server.

### 3) Local SSH Server

We will start a local SSH server running for the duration of the program using `Teleport`.
The server will be listening on an arbitrary port.
We will call this the `$LocalPort`.
This will be the TCP port used to SSHfs mount the local drive from the remote pod.


### 4) Reverse SSH Tunnel

In order to mount a local drive remotely, we will need to build a reverse SSH tunnel.
On a `kamp run` we will kick off a concurrent process that will use Helm to deploy these manifests.
Once the SSH server is running, and publicly accessible we will create a reverse SSH tunnel to the server.
The tunnel will map `$LocalPort` to `127.0.0.1` on the pod.
The pod (that ships with SSHfs) will now be able to remotely mount the your local filesystem in Kubernetes.

### 5) NFS mounting to your container

At runtime a user specified their own arbitrary container they would like to use.
We will now NFS share the SSHfs volume that is mounted on the pod.
The users container will be able to NFS mount this volume via `127.0.0.1`.
The container was originally deployed with the NFS volume defined.

### 6) Attach to your container

Now that we have successfully mounted a volume from your local workstation to an arbitrary container of your choosing directly in Kubernetes, the next step is to attach to the container.
Kamp will (by default) attach a user to their container using `/bin/bash`.


