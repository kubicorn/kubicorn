# ------------------------------------------------------------------------------------------------------------------------
# We are explicitly not using a templating language to inject the values as to encourage the user to limit their
# use of templating logic in these files. By design all injected values should be able to be set at runtime,
# and the shell script real work. If you need conditional logic, write it in bash or make another shell script.
# ------------------------------------------------------------------------------------------------------------------------

# Disabling SELinux is not recommended and will be fixed later.
sudo sed -i 's/^SELINUX=.*/SELINUX=disabled/g' /etc/sysconfig/selinux
sudo setenforce 0

sudo rpm --import https://packages.cloud.google.com/yum/doc/yum-key.gpg
sudo rpm --import https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg

sudo sh -c 'cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=http://yum.kubernetes.io/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg
       https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOF'

sudo yum makecache -y
sudo yum install -y \
     docker \
     socat \
     ebtables \
     kubelet \
     kubeadm \
     cloud-utils \
     epel-release

# jq needs its own special yum install as it depends on epel-release
sudo yum install -y jq

# Has to be configured before starting kubelet, or kubelet has to be restarted to pick up changes
sudo sh -c 'cat <<EOF > /etc/systemd/system/kubelet.service.d/20-cloud-provider.conf
[Service]
Environment="KUBELET_EXTRA_ARGS=--cloud-provider=aws"
EOF'

sudo systemctl enable docker
sudo systemctl enable kubelet.service
sudo systemctl start docker

PUBLICIP=$(ec2metadata --public-ipv4 | cut -d " " -f 2)
PRIVATEIP=$(ip addr show dev eth0 | awk '/inet / {print $2}' | cut -d"/" -f1)
TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.values.itemMap.INJECTEDTOKEN')
PORT=$(cat /etc/kubicorn/cluster.json | jq -r '.values.itemMap.INJECTEDPORT | tonumber')
# Necessary for joining a cluster with the AWS information
HOSTNAME=$(hostname -f)

# Required by kubeadm
sudo sysctl -w net.bridge.bridge-nf-call-iptables=1
sudo sysctl -p

cat << EOF  > "/etc/kubicorn/kubeadm-config.yaml"
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
cloudProvider: aws
token: ${TOKEN}
nodeName: ${HOSTNAME}
api:
  advertiseAddress: ${PUBLICIP}
  bindPort: ${PORT}
apiServerCertSANs:
- ${PUBLICIP}
- ${HOSTNAME}
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

kubectl apply \
    -f  https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.8/cluster/addons/storage-class/aws/default.yaml \
    --kubeconfig /etc/kubernetes/admin.conf

# Default centos user
mkdir -p /home/centos/.kube
cp /etc/kubernetes/admin.conf /home/centos/.kube/config
chown -R centos:centos /home/centos/.kube

