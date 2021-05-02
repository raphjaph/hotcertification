#!/usr/bin/env bash

export HOTSTUFF_LOG=info

trap 'trap - SIGTERM && kill -- -$$' SIGINT SIGTERM EXIT

bin='go run cmd/certificationserver/*.go'

$bin --id 1 --thresholdkey keys/p1.thresholdkey --privkey keys/p1.key &
$bin --id 2 --thresholdkey keys/p2.thresholdkey --privkey keys/p2.key 2> logs/2.out &
$bin --id 3 --thresholdkey keys/p3.thresholdkey --privkey keys/p3.key 2> logs/3.out &
$bin --id 4 --thresholdkey keys/p4.thresholdkey --privkey keys/p4.key 2> logs/4.out &

if [ "$1" = "kill" ]; then
	sleep 5s
	kill $!
fi

wait; wait; wait; wait;
