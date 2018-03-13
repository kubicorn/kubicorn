# ------------------------------------------------------------------------------------------------------------------------
# We are explicitly not using a templating language to inject the values as to encourage the user to limit their
# use of templating logic in these files. By design all injected values should be able to be set at runtime,
# and the shell script real work. If you need conditional logic, write it in bash or make another shell script.
# ------------------------------------------------------------------------------------------------------------------------

# Specify the Kubernetes version to use
KUBERNETES_VERSION="1.9.0"

apt-get update -y
apt-get install -y \
    socat \
    cloud-utils \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common \
    jq

curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg | sudo apt-key add -

add-apt-repository \
   "deb https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
   $(lsb_release -cs) \
   stable"

apt-get update && apt-get install -y docker-ce=$(apt-cache madison docker-ce | grep 17.03 | head -1 | awk '{print $3}')

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

cat << EOF > "/etc/apt/sources.list.d/kubernetes.list"
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF

systemctl enable docker
systemctl start docker

apt-get update -y
apt-get install -y \
    kubelet \
    kubeadm=${KUBERNETES_VERSION}-00 \
    kubectl

PUBLICIP=$(ec2metadata --public-ipv4 | cut -d " " -f 2)
PRIVATEIP=$(ec2metadata --local-ipv4 | cut -d " " -f 2)
TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDTOKEN')
PORT=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDPORT | tonumber')

mkdir -p /etc/kubicorn

cat << EOF  > "/etc/kubicorn/kubeadm-config.yaml"
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
token: ${TOKEN}
kubernetesVersion: ${KUBERNETES_VERSION}
api:
  advertiseAddress: ${PUBLICIP}
  bindPort: ${PORT}
apiServerCertSANs:
- ${PUBLICIP}
- ${PRIVATEIP}
authorizationModes:
- Node
- RBAC
EOF

kubeadm reset
kubeadm init --config /etc/kubicorn/kubeadm-config.yaml

# Thanks Kelsey :)
kubectl apply \
  -f http://docs.projectcalico.org/v2.3/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml \
  --kubeconfig /etc/kubernetes/admin.conf

mkdir -p /home/admin/.kube
cp /etc/kubernetes/admin.conf /home/admin/.kube/config
chown -R admin:admin /home/admin/.kube
