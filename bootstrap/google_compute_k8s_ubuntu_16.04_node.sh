# ------------------------------------------------------------------------------------------------------------------------
# We are explicitly not using a templating language to inject the values as to encourage the user to limit their
# use of templating logic in these files. By design all injected values should be able to be set at runtime,
# and the shell script real work. If you need conditional logic, write it in bash or make another shell script.
# ------------------------------------------------------------------------------------------------------------------------

# Specify the Kubernetes version to use.
KUBERNETES_VERSION="1.9.2"
KUBERNETES_CNI="0.6.0"

# Obtain metadata.
PRIVATEIP=`curl --retry 5 -sfH "Metadata-Flavor: Google" "http://metadata/computeMetadata/v1/instance/network-interfaces/0/ip"`
PUBLICIP=`curl --retry 5 -sfH "Metadata-Flavor: Google" "http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip"`
HOSTNAME=`curl --retry 5 -sfH "Metadata-Flavor: Google" "http://metadata/computeMetadata/v1/instance/name"`
echo $PRIVATEIP > /tmp/.ip

# Add APT repository and GPG key.
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
touch /etc/apt/sources.list.d/kubernetes.list
sh -c 'echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list'

# Install packages.
apt-get update -y
apt-get install -y \
    socat \
    ebtables \
    docker.io \
    apt-transport-https \
    kubelet=${KUBERNETES_VERSION}-00 \
    kubeadm=${KUBERNETES_VERSION}-00 \
    kubernetes-cni=${KUBERNETES_CNI}-00 \
    cloud-utils \
    jq

# Enable and start Docker.
systemctl enable docker
systemctl start docker

# Parse kubicorn configuration file.
TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDTOKEN')
MASTER=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDMASTER')

# Join node a cluster.
kubeadm reset
kubeadm join --node-name ${HOSTNAME} --token ${TOKEN} ${MASTER} --discovery-token-unsafe-skip-ca-verification
