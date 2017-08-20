#!/usr/bin/env bash
set -e
cd ~

# ------------------------------------------------------------------------------------------------------------------------
# These values are injected into the script. We are explicitly not using a templating language to inject the values
# as to encourage the user to limit their use of templating logic in these files. By design all injected values should
# be able to be set at runtime, and the shell script real work. If you need conditional logic, write it in bash
# or make another shell script.
#
# Required parameters 
TOKEN="INJECTEDTOKEN"
MASTER="INJECTEDMASTER"
PORT="INJECTEDPORT"
CLOUDPROVIDER="INJECTEDCLOUDPROVIDER"
ARCH="INJECTEDARCH"  # amd, amd64, arm, arm64
OS="INJECTEDOS" # ubuntu, centos, coreOS, fedora, RHEL
NODETYPE="INJECTEDNODETYPE"  # or MASTER

# Optional parameters
BIN_DIR=${BIN_DIR:-/usr/bin}
INSTALL_K8S_VERSION=${K8S_VERSION:-1.7.0}  # defaults to latest release
PRIVATEIP="PRIVATEIP"
PUBLICIP="PUBLICIP"
ROOTFS="/"
K8S_URL=${K8S_URL:-https://dl.k8s.io/}
CNI_URL=${CNI_URL:-https://storage.googleapis.com/kubernetes-release/network-plugins/cni-${ARCH}-${INSTALL_CNI_RELEASE}.tar.gz}

# ------------------------------------------------------------------------------------------------------------------------

# Cloud Provider specific settings
echo "Cloud provider is ${CLOUDPROVIDER}"
if [[ ${CLOUDPROVIDER} == "aws" ]] || [[ ${CLOUDPROVIDER} == "amazon" ]]; then
    PUBLICIP=$(ec2metadata --public-ipv4 | cut -d " " -f 2)
    PRIVATEIP=$(ip addr show dev eth0 | awk '/inet / {print $2}' | cut -d"/" -f1)

elif [[ ${CLOUDPROVIDER} == "gce" ]] || [[ ${CLOUDPROVIDER} == "google" ]]; then
    PRIVATEIP=`curl --retry 5 -sfH "Metadata-Flavor: Google" "http://metadata/computeMetadata/v1/instance/network-interfaces/0/ip"`
    PUBLICIP=`curl --retry 5 -sfH "Metadata-Flavor: Google" "http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip"`

elif [[ ${CLOUDPROVIDER} == "do" ]] || [[ ${CLOUDPROVIDER} == "digitalocean" ]]; then
    PRIVATEIP=$(ip addr show dev tun0 | awk '/inet / {print $2}' | cut -d"/" -f1)
    PUBLICIP=$(curl -sSL http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address)

elif [[ ${CLOUDPROVIDER} == "azure" ]]; then
    echo "Cloud provider is ${CLOUDPROVIDER}"
    exit 255 # not implemented yet
else 
    echo "Error : Cloud Provider unknown ${CLOUDPROVIDER}"
    exit 255
fi

if [[ ${PRIVATEIP} == "PRIVATEIP" ]]; then 
    echo "Error : unable to find private IP"
fi
if [[ ${PUBLICIP} == "PUBLICIP" ]]; then 
    echo "Error : unable to find public IP"
fi

echo $PRIVATEIP > /tmp/.ip

if [[ ${OS} == "ubuntu" ]]; then 
   
    curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
    touch /etc/apt/sources.list.d/kubernetes.list
    sh -c 'echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list'

    apt-get update -y
    apt-get install -y \
        socat \
        ebtables \
        docker.io \
        apt-transport-https \
        kubelet=${INSTALL_K8S_VERSION}-00\
        kubeadm=1.7.0-00  # get latest Kubeadm

elif [[ ${OS} == "centos" ]]; then 
    
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
    sudo sudo yum install -y \
        docker \
        socat \
        ebtables \
        kubelet=${INSTALL_K8S_VERSION}-00 \
        kubeadm=1.7.0-00 \ # get latest Kubeadm
        cloud-utils
  
     # Required by kubeadm
    sudo sysctl -w net.bridge.bridge-nf-call-iptables=1
    sudo sysctl -p

else 

    echo "Error : Unknown host OS ${OS}"
    exit 255
fi

# start Kublet service
systemctl daemon-reload
systemctl enable kubelet
systemctl start kubelet

#start Docker
sudo systemctl enable docker
sudo systemctl start docker

sudo kubeadm reset

# MASTER settings 
if [[ ${NODETYPE} == "MASTER" ]] || [[ ${NODETYPE} == "master" ]]; then
  
    kubeadm init --apiserver-bind-port ${PORT} --token ${TOKEN}  --apiserver-advertise-address ${PUBLICIP} --apiserver-cert-extra-sans ${PUBLICIP} ${PRIVATEIP} --kubernetes-version v${INSTALL_K8S_VERSION}

    kubectl apply \
        -f http://docs.projectcalico.org/v2.3/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml \
        --kubeconfig /etc/kubernetes/admin.conf

    # Root
    mkdir -p ~/.kube
    cp /etc/kubernetes/admin.conf ~/.kube/config

elif [[ ${NODETYPE} == "NODE" ]] || [[ ${NODETYPE} == "node" ]]; then
    sudo -E kubeadm join --token ${TOKEN} ${MASTER}

else 
    echo "Error : Unknown NODETYPE ${NODETYPE}"
fi