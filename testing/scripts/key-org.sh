#!/bin/bash

dir=$1
n=$2

cd $dir
for i in $(seq 1 $n)
do
    mkdir $i;
    cp ../hotcertification.yml $i
    mv n$i.thresholdkey n$i.key $i
    cp *.pub *.crt $i
done

rm *.pub *.crt

