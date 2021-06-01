#!/bin/bash

num_nodes=$1

for i in $(seq 1 $num_nodes)
do
docker stop n$i
done


docker container prune

