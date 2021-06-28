#!/bin/bash

if [[ $# -lt 1 ]] ; then
    echo 'please provide key directory (relative path) and the number of servers to start'
    exit 0
fi

num_nodes=$1
threshold=$(echo "$num_nodes - ($num_nodes - 1) / 3" | bc)

# Generate a config file for docker deployment (same network)
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
docker run --name "keygen" -v "$(pwd):/home" raphasch/hotcertification keygen -n $num_nodes -t $threshold --key-size 512 keys

# Structuring keys into seperate directories
cd keys
for i in $(seq 1 $num_nodes)
do
    mkdir $i;
    cp ../hotcertification.toml $i
    mv n$i.thresholdkey n$i.key $i
    cp *.pub *.crt $i
done
rm *.pub *.crt

cd ..

# test file for csr-size variable
mkdir measurements
truncate -s 100B measurements/100B.info
truncate -s 1K measurements/1K.info
truncate -s 1M measurements/1M.info
truncate -s 100M measurements/100M.info

# create first container that logs and which the client connects to.
docker run -d --name n1 --publish "8081:8081" --volume "$(pwd)/keys/1:/home" --env "HOTSTUFF_LOG=info" --network "hotcertification" raphasch/hotcertification certificationserver --id 1 --thresholdkey n1.thresholdkey --privkey n1.key &

for i in $(seq 2 $num_nodes)
do
docker run -d --name n$i -v "$(pwd)/keys/$i:/home" --network "hotcertification" raphasch/hotcertification certificationserver --id $i --thresholdkey n$i.thresholdkey --privkey n$i.key &
done
