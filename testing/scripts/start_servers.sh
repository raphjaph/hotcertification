#!/bin/bash

num_nodes=$1

# create first container that logs and which the client connects to.
docker run -d --name n1 --publish "8081:8081" --volume "/root/hotcertification/testing/keys/1:/home" --env "HOTSTUFF_LOG=info" --network "hotcertification" raphasch/hotcertification certificationserver --id 1 --thresholdkey n1.thresholdkey --privkey n1.key &

for i in $(seq 2 $num_nodes)
do
docker run -d --name n$i -v "/root/hotcertification/testing/keys/$i:/home" --network "hotcertification" raphasch/hotcertification certificationserver --id $i --thresholdkey n$i.thresholdkey --privkey n$i.key &
done