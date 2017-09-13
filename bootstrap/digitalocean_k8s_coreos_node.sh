#!/usr/bin/env bash
set -e
cd ~

# ------------------------------------------------------------------------------------------------------------------------
# These values are injected into the script. We are explicitly not using a templating language to inject the values
# as to encourage the user to limit their use of templating logic in these files. By design all injected values should
# be able to be set at runtime, and the shell script real work. If you need conditional logic, write it in bash
# or make another shell script.
#
#
TOKEN="INJECTEDTOKEN"
MASTER="INJECTEDMASTER"
# ------------------------------------------------------------------------------------------------------------------------

K8S_VERSION=v1.7.0

curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl > /opt/bin/kubectl
curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubeadm > /opt/bin/kubeadm

sudo systemctl enable docker
sudo systemctl start docker

sudo -E kubeadm reset
sudo -E kubeadm join --token ${TOKEN} ${MASTER}
