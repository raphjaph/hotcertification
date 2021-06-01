#!/usr/bin/env bash

export HOTSTUFF_LOG=info

trap 'trap - SIGTERM && kill -- -$$' SIGINT SIGTERM EXIT

bin='cmd/certificationserver/certserver'

$bin --id 1 --thresholdkey keys/n1.thresholdkey --privkey keys/n1.key &
$bin --id 2 --thresholdkey keys/n2.thresholdkey --privkey keys/n2.key 2> logs/2.out &
$bin --id 3 --thresholdkey keys/n3.thresholdkey --privkey keys/n3.key 2> logs/3.out &
$bin --id 4 --thresholdkey keys/n4.thresholdkey --privkey keys/n4.key 2> logs/4.out &

if [ "$1" = "kill" ]; then
	sleep 5s
	kill $!
fi

wait; wait; wait; wait;
