#!/bin/bash

if [[ $# -eq 0 ]] ; then
    echo 'please provide a directory for where the measurments should be saved and the amount of clients started'
    exit 0
fi
measure_dir=$1
num_clients=$2

for i in $(seq 1 $num_clients)
do
docker run -d --name client$i -v "$(pwd)/$measure_dir:/home" --network "hotcertification" raphasch/hotcertification client client$i.crt --server-addr "n$i:8081" --file $i.csv
done

docker container prune