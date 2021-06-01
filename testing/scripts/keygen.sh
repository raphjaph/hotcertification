#!/bin/bash

# This script starts a docker container to generate the keys (trusted setup phase)
# The keys are then stored on host machine in the specified folder
# This script should also generate a config file 
# Then config file and keys are organized into node folders


docker run --name "keygen" raphasch/hotcertification keygen -n 4 -t 3 key-size 2048 keys
docker cp keygen:/home/keys .


