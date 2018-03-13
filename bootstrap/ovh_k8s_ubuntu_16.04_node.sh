# ------------------------------------------------------------------------------------------------------------------------
# We are explicitly not using a templating language to inject the values as to encourage the user to limit their
# use of templating logic in these files. By design all injected values should be able to be set at runtime,
# and the shell script real work. If you need conditional logic, write it in bash or make another shell script.
# ------------------------------------------------------------------------------------------------------------------------

# Specify the Kubernetes version to use.
KUBERNETES_VERSION="1.9.2"
KUBERNETES_CNI="0.6.0"

# Acquire private IP address
cat << EOF >> "/etc/network/interfaces.d/50-cloud-init.cfg"
auto ens4
iface ens4 inet dhcp
post-up ip route add 10.96.0.0/12 dev ens4
pre-down ip route del 10.96.0.0/12 dev ens4
EOF
systemctl restart networking

# Obtain IP addresses.
HOSTNAME=$(curl -s http://169.254.169.254/latest/meta-data/hostname)
PUBLICIP=$(ip -f inet -o addr show ens3|cut -d\  -f 7 | cut -d/ -f 1)
PRIVATEIP=$(ip -f inet -o addr show ens4|cut -d\  -f 7 | cut -d/ -f 1)

# Add Kubernetes repository.
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

# Specify node IP for kubelet.
echo "Environment=\"KUBELET_EXTRA_ARGS=--node-ip=${PRIVATEIP}\"" >> /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
systemctl daemon-reload
systemctl restart kubelet

# Parse kubicorn configuration file.
TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDTOKEN')
MASTER=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDMASTER')

# Join node a cluster.
kubeadm reset
kubeadm join --node-name ${HOSTNAME} --token ${TOKEN} ${MASTER} --discovery-token-unsafe-skip-ca-verification
