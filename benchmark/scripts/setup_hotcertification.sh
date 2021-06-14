#!/bin/bash

# This script starts a docker container to generate the keys (trusted setup phase)
# The keys are then stored on host machine in the specified folder
# This script should also generate a config file 
# Then config file and keys are organized into node folders


if [[ $# -lt 2 ]] ; then
    echo 'please provide a directory for where the keys should be saved and the number of nodes'
    exit 0
fi

dir=$1
num_nodes=$2
threshold=$(echo "$num_nodes - ($num_nodes - 1) / 3" | bc)

# Generate a config file for docker deplyment (same network)
cat <<EOF >> hotcertification.toml
# For TLS 
root-ca = "root.crt"

# HotStuff config options
pacemaker = "round-robin"
leader-id = 1
view-timeout = 100

# Size of RSA key that certificates are signed with
# must be same size as set when generating keys with keygen executable
key-size = 512


# This is the information that each replica is given about the other replicas
EOF

for i in $(seq 1 $num_nodes)
do
cat <<EOF >> hotcertification.toml
[[nodes]]
id = $i
pubkey = "n$i.key.pub"
tls-cert = "n$i.crt"
client-srv-address = "n$i:8081"
replication-srv-address = "n$i:13371"
signing-srv-address = "n$i:23371"

EOF
done

# Generating keys
docker run --name "keygen" -v "$(pwd):/home" raphasch/hotcertification keygen -n $num_nodes -t $threshold --key-size 512 $dir

# Structuring keys into seperate directories
cd $dir
for i in $(seq 1 $num_nodes)
do
    mkdir $i;
    cp ../hotcertification.toml $i
    mv n$i.thresholdkey n$i.key $i
    cp *.pub *.crt $i
done
rm *.pub *.crt

# Removing container
docker container prune -f
