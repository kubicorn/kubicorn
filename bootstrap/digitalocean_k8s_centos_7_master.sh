# ------------------------------------------------------------------------------------------------------------------------
# We are explicitly not using a templating language to inject the values as to encourage the user to limit their
# use of templating logic in these files. By design all injected values should be able to be set at runtime,
# and the shell script real work. If you need conditional logic, write it in bash or make another shell script.
# ------------------------------------------------------------------------------------------------------------------------

# Specify the Kubernetes version to use.
KUBERNETES_VERSION="1.9.2"
KUBERNETES_CNI="0.6.0"

# Import GPG keys and add repository entries for Kuberenetes.
rpm --import https://packages.cloud.google.com/yum/doc/yum-key.gpg
rpm --import https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg

cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=http://yum.kubernetes.io/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
       https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF

# Install packages.
yum makecache -y
yum install -y \
     docker \
     socat \
     ebtables \
     kubelet-${KUBERNETES_VERSION}-0 \
     kubeadm-${KUBERNETES_VERSION}-0 \
     kubernetes-cni-${KUBERNETES_CNI}-0 \
     cloud-utils \
     epel-release

# "jq" depends on epel-release, so it needs its own yum install command.
sudo yum install -y jq

# Enable Docker and Kubelet services.
systemctl enable docker
systemctl enable kubelet
systemctl start docker

# Obtain Droplet IP addresses.
HOSTNAME=$(curl -s http://169.254.169.254/metadata/v1/hostname)
PRIVATEIP=$(curl -s http://169.254.169.254/metadata/v1/interfaces/private/0/ipv4/address)
PUBLICIP=$(curl -s http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address)
VPNIP=$(ip addr show dev tun0 | awk '/inet / {print $2}' | cut -d"/" -f1)
echo $VPNIP > /tmp/.ip

# Specify node IP for kubelet.
echo "Environment=\"KUBELET_EXTRA_ARGS=--node-ip=${PUBLICIP}\"" >> /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
systemctl daemon-reload
systemctl restart kubelet

# Parse Kubicorn configuration file.
TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDTOKEN')
PORT=$(cat /etc/kubicorn/cluster.json | jq -r '.clusterAPI.spec.providerConfig' | jq -r '.values.itemMap.INJECTEDPORT | tonumber')

# Required by kubeadm.
sysctl -w net.bridge.bridge-nf-call-iptables=1
sysctl -p

# Create kubeadm configuration file.
touch /etc/kubicorn/kubeadm-config.yaml
cat << EOF  > "/etc/kubicorn/kubeadm-config.yaml"
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
token: ${TOKEN}
kubernetesVersion: ${KUBERNETES_VERSION}
nodeName: ${HOSTNAME}
api:
  advertiseAddress: ${PUBLICIP}
  bindPort: ${PORT}
apiServerCertSANs:
- ${PRIVATEIP}
- ${PUBLICIP}
- ${HOSTNAME}
authorizationModes:
- Node
- RBAC
EOF

# Initialize cluster.
kubeadm reset
kubeadm init --config /etc/kubicorn/kubeadm-config.yaml

# Weave CNI plugin.
curl -SL "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')&env.IPALLOC_RANGE=172.16.6.64/27" \
| kubectl apply --kubeconfig /etc/kubernetes/admin.conf -f -

mkdir -p /root/.kube
cp /etc/kubernetes/admin.conf /root/.kube/config
chown -R root:root /root/.kube
