#!/bin/bash

node=bitcoin

echo "Allocating $node"
pos allocations allocate $node

pos nodes image $node debian-buster

echo "Restarting machine with new OS"
pos nodes reset $node

echo "Copying repo to machine"
scp -q -r /home/schleith/hotcertification $node:~

echo "Installing docker on machine"
pos commands launch $node -- sh hotcertification/benchmark/scripts/install-docker-debian.sh

echo "Pulling hotcertification image from docker hub"
pos commands launch $node -- docker pull raphasch/hotcertification

echo "Creating docker network for hotcertification cluster"
pos commands launch $node -- docker network create hotcertification
