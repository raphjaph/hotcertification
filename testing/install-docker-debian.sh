#!/bin/bash

apt update
apt install -y apt-transport-https ca-certificates curl gnupg2 software-properties-common

curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"

apt update
apt-cache policy docker-ce
apt install -y docker-ce

#Workaround for an error; needs to know that system is using a RAM disk
mkdir -p /etc/systemd/system/docker.service.d
cat >/etc/systemd/system/docker.service.d/10-ramdisk.conf <<EOF
[Service]
Environment=DOCKER_RAMDISK=true
EOF
systemctl daemon-reload
systemctl restart docker
