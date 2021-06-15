#!/bin/bash

if [[ $# -lt 2 ]] ; then
    echo 'please provide a directory for where the measurments should be saved and the amount of clients started'
    exit 0
fi
dest=$1
num_clients=$2

for i in $(seq 1 $num_clients)
do
docker run -d --name client$i -v "$(pwd)/$dest:/home" --network "hotcertification" raphasch/hotcertification benchmark m$i.csv --server-addr "n$i:8081"
done

docker container prune -f