#!/bin/bash

if [[ $# -lt 3 ]] ; then
    echo 'Please provide the scenario defintion, the validation info mock file and the amount of clients to start'
    exit 0
fi

scenario=$1
file=$2
num_clients=$3

for i in $(seq 1 $num_clients)
do
docker run -d --name b$i -v "$(pwd)/measurements:/home" --network "hotcertification" raphasch/hotcertification benchmark --scenario $scenario --file $file --server-addr "n$i:8081" $dir/m$i.csv
done
