#!/bin/bash

# This script starts a docker container to generate the keys (trusted setup phase)
# The keys are then stored on host machine in the specified folder
# This script should also generate a config file 
# Then config file and keys are organized into node folders

docker run  --name "keygen" -v "/root/hotcertification/testing:/home" raphasch/hotcertification keygen -n 4 -t 3 --key-size 512 keys
docker container prune
