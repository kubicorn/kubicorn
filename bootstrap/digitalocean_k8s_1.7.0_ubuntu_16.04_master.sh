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
NAME="INJECTEDNAME"
#
#
# Optionally defined parameters used for the VPN generation. These are not injected, but you are welcome to change them
# for your use case.
#
#
VPNCONFIG="
# --------------------------------------
# Injected by Kubicorn
#
export KEY_COUNTRY='US'
export KEY_PROVINCE='CO'
export KEY_CITY='Boulder'
export KEY_ORG='Boulder'
export KEY_EMAIL='me@kubicorn.io'
export KEY_OU='kubicornz'
# --------------------------------------"
#
# ------------------------------------------------------------------------------------------------------------------------

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
touch /etc/apt/sources.list.d/kubernetes.list
sh -c 'echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list'

apt-get update -y
apt-get install -y \
    socat \
    ebtables \
    docker.io \
    apt-transport-https \
    kubelet \
    kubeadm=1.7.0-00 \
    cloud-utils


systemctl enable docker
systemctl start docker

PUBLICIP=$(curl ifconfig.me)
PRIVATEIP=$(ifconfig | grep -A 1 eth0 | grep inet | cut -d ":" -f 2 | cut -d " " -f 1 | xargs)


kubeadm reset
kubeadm init --apiserver-bind-port ${PORT} --token ${TOKEN}  --apiserver-advertise-address ${PUBLICIP} --apiserver-cert-extra-sans ${PUBLICIP} ${PRIVATEIP}


kubectl apply \
  -f http://docs.projectcalico.org/v2.3/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml \
  --kubeconfig /etc/kubernetes/admin.conf

# Root
mkdir -p ~/.kube
cp /etc/kubernetes/admin.conf ~/.kube/config

# VPN Mesh
apt-get update -y && apt-get install openvpn easy-rsa -y
mkdir /etc/openvpn/easy-rsa/
echo $VPNCONFIG >> /etc/openvpn/easy-rsa/vars
cd /etc/openvpn/easy-rsa/
source vars
./clean-all
./pkitool --initca ${NAME}
./pkitool --server ${NAME}

cd /etc/openvpn/easy-rsa/keys/
cp ${NAME}.crt ${NAME}.key ca.crt ca.key dh2048.pem /etc/openvpn/

# Generate client keys
cd /etc/openvpn/easy-rsa/
source vars
./pkitool ${NAME}

# Client Keys to be copied over to the clients
#/etc/openvpn/ca.crt
#/etc/openvpn/easy-rsa/keys/${NAME}.crt
#/etc/openvpn/easy-rsa/keys/${NAME}.key

# Start VPN Server
cp /usr/share/doc/openvpn/examples/sample-config-files/server.conf.gz /etc/openvpn/
gzip -d /etc/openvpn/server.conf.gz

sed -i "s|cert server.crt|cert ${NAME}.crt|g" /etc/openvpn/server.conf
sed -i "s|key server.key|key ${NAME}.key|g" /etc/openvpn/server.conf
sed -i "s|:client-to-client|client-to-config|g" /etc/openvpn/server.conf

systemctl enable openvpn
systemctl start openvpn