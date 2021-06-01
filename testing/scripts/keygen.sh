#!/bin/bash

# This script starts a docker container to generate the keys (trusted setup phase)
# The keys are then stored on host machine in the specified folder
# This script should also generate a config file 
# Then config file and keys are organized into node folders

dir=$1
num_nodes=$2

docker run --name "keygen" -v "/root/hotcertification/testing:/home" raphasch/hotcertification keygen -n $num_nodes -t 3 --key-size 512 $dir

cd $dir
for i in $(seq 1 $num_nodes)
do
    mkdir $i;
    cp ../hotcertification.yml $i
    mv n$i.thresholdkey n$i.key $i
    cp *.pub *.crt $i
done

rm *.pub *.crt

docker container prune
