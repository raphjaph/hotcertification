#!/usr/bin/env bash

export HOTCERTIFICATION_LOG=debug

trap 'trap - SIGTERM && kill -- -$$' SIGINT SIGTERM EXIT

bin='go run cmd/certificationserver/*.go'

$bin --id 1 --thresholdkey keys/p1.thresholdkey &
$bin --id 2 --thresholdkey keys/p2.thresholdkey 2> logs/2.out &
$bin --id 3 --thresholdkey keys/p3.thresholdkey 2> logs/3.out &

if [ "$1" = "kill" ]; then
	sleep 5s
	kill $!
fi

wait; wait; wait; wait;
