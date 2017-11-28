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
    jq

systemctl enable docker
systemctl start docker

# must disable swap for kubelet to work
swapoff -a


TOKEN=$(cat /etc/kubicorn/cluster.json | jq -r '.values.itemMap.INJECTEDTOKEN')
MASTER=$(cat /etc/kubicorn/cluster.json | jq -r '.values.itemMap.INJECTEDMASTER')

systemctl daemon-reload
systemctl restart kubelet.service

kubeadm reset
kubeadm join --token ${TOKEN} ${MASTER}
