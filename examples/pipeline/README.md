# Example of using Kubicorn to deploy clusters with a pipeline system like Jenkins

### Load custom bootstrap script

```bash

# Make a directory to store local bootstrap scripts in
mkdir -p ~/.kubicorn/bootstrap/

# Make a directory to store local state files in
mkdir -p ~/.kubicorn/state

# Curl down example bootstrap scripts, feel free to modify or bring your own
curl https://raw.githubusercontent.com/kubicorn/bootstrap/master/amazon_k8s_ubuntu_16.04_master.sh -o ~/.kubicorn/bootstrap/amazon_k8s_ubuntu_16.04_master.sh
curl https://raw.githubusercontent.com/kubicorn/bootstrap/master/amazon_k8s_ubuntu_16.04_node.sh -o ~/.kubicorn/bootstrap/amazon_k8s_ubuntu_16.04_node.sh
```

Now we can create a cluster in AWS using these bootstrap scripts and some other commonly overridden fields as well.


```bash
ID=$(date +%s)
NAME=my-pipeline-cluster-${ID}
kubicorn create $NAME\
    -S ${HOME}/.kubicorn/state \
    -p aws \
    -C location=us-west-1 \
    -M serverPool.image="my-ami-123" \
    -M serverPool.bootstrapScripts[0]="${HOME}/.kubicorn/bootstrap/amazon_k8s_ubuntu_16.04_master.sh" \
    -M serverPool.subnets[0].zone="us-west-1a" \
    -N serverPool.image="my-ami-456" \
    -N serverPool.subnets[0].zone="us-west-1b" \
    -N serverPool.bootstrapScripts[0]="${HOME}/.kubicorn/bootstrap/amazon_k8s_ubuntu_16.04_master.sh"
```