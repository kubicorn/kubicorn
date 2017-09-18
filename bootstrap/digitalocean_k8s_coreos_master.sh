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
PORT="INJECTEDPORT"
#
#
# ------------------------------------------------------------------------------------------------------------------------

K8S_VERSION=v1.7.0

mkdir -p /opt/cni /opt/bin
curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl > /opt/bin/kubectl
curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubeadm > /opt/bin/kubeadm
chmod +x /opt/bin/kubectl /opt/bin/kubeadm

systemctl enable docker
systemctl start docker

PUBLICIP=$(curl ifconfig.me)
PRIVATEIP=$(ip addr show dev eth0 | awk '/inet / {print $2}' | awk 'FNR == 2 {print}' | cut -d"/" -f1)
echo $PRIVATEIP > /tmp/.ip

kubeadm reset
kubeadm init --apiserver-bind-port ${PORT} --token ${TOKEN}  --apiserver-advertise-address ${PUBLICIP} --apiserver-cert-extra-sans ${PUBLICIP} ${PRIVATEIP}


kubectl apply \
  -f http://docs.projectcalico.org/v2.3/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml \
  --kubeconfig /etc/kubernetes/admin.conf

# Root
mkdir -p /home/ubuntu/.kube
cp /etc/kubernetes/admin.conf /home/ubuntu/.kube/config
chown -R ubuntu:ubuntu /home/ubuntu/.kube
