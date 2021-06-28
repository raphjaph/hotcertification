#!/bin/bash

# [num-nodes, csr-size, adversary-type, adversary-fraction]
scenario="4,100,none,0"
file="100B.info"
num_clients=4
num_requests=$(echo "1000 / $num_clients" | bc)

for i in $(seq 1 $num_clients)
do
docker run -d --name b$i -v "$(pwd)/measurements:/home" --network "hotcertification" raphasch/hotcertification benchmark -n $num_requests --scenario $scenario --file $file --server-addr "n$i:8081" m$scenario-$i.csv
done
