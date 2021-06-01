#!/bin/bash

echo "Allocating bitcoingold"
pos allocations allocate bitcoingold

pos nodes image bitcoingold debian-buster

echo "Restarting machine with new OS"
pos nodes reset bitcoingold

echo "Copying repo to machine"
scp -q -r /home/schleith/hotcertification bitcoingold:~

echo "Installing docker on machine"
pos commands launch bitcoingold -- sh hotcertification/testing/install-docker-debian.sh

echo "Pulling hotcertification image from docker hub"
pos commands launch bitcoingold -- docker pull raphasch/hotcertification

