#!/usr/bin/env bash

export HOTSTUFF_LOG=debug

trap 'trap - SIGTERM && kill -- -$$' SIGINT SIGTERM EXIT

bin='go run server/*.go'

$bin --self-id 1 --privkey keys/r1.key --batch-size 1 --print-commands  2> logs/1.out &
$bin --self-id 2 --privkey keys/r2.key --batch-size 1 2> logs/2.out &
$bin --self-id 3 --privkey keys/r3.key --batch-size 1 2> logs/3.out &
$bin --self-id 4 --privkey keys/r4.key --batch-size 1 2> logs/4.out &

if [ "$1" = "kill" ]; then
	sleep 5s
	kill $!
  killall -9 certificationserver
fi

wait; wait; wait; wait;
