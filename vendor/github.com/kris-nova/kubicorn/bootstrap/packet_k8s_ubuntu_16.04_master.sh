# ------------------------------------------------------------------------------------------------------------------------
# We are explicitly not using a templating language to inject the values as to encourage the user to limit their
# use of templating logic in these files. By design all injected values should be able to be set at runtime,
# and the shell script real work. If you need conditional logic, write it in bash or make another shell script.
# ------------------------------------------------------------------------------------------------------------------------

# order is important:
# 1. update
# 2. install apt-transport-https
# 3. add kubernetes repos to list
# 4. update again
# 5. install
apt-get update -y
apt-get install -y apt-transport-https

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
touch /etc/apt/sources.list.d/kubernetes.list
sh -c 'echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list'

apt-get update -y

apt-get install -y \
    socat \
    ebtables \
    docker.io \
    apt-transport-https \
    kubelet \
    kubeadm=1.7.0-00 \
    cloud-utils \
    jq


systemctl enable docker
systemctl start docker

# must disable swap for kubelet to work
swapoff -a

PUBLICIP=$(curl --silent  https://metadata.packet.net/metadata | jq '.network.addresses[] | select(.address_family == 4 and .public == true) .address')
PRIVATEIP=$(curl --silent  https://metadata.packet.net/metadata | jq '.network.addresses[] | select(.address_family == 4 and .public == false) .address')
TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.values.itemMap.INJECTEDTOKEN')
PORT=$(cat /etc/kubicorn/cluster.json | jq -r '.kubernetesAPI.port | tonumber')

kubeadm reset
kubeadm init --apiserver-bind-port ${PORT} --token ${TOKEN}  --apiserver-advertise-address ${PUBLICIP} --apiserver-cert-extra-sans ${PUBLICIP} ${PRIVATEIP}

# Thanks Kelsey :)
kubectl apply \
  -f http://docs.projectcalico.org/v2.3/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml \
  --kubeconfig /etc/kubernetes/admin.conf

mkdir -p /root/.kube
cp /etc/kubernetes/admin.conf /root/.kube/config
