#!/bin/bash

dir=$1
n=$2

cd $dir
for i in $(seq 1 $n)
do
    mkdir $i;
    cp ../hotcertification.yml $i
done

mv n1.thresholdkey n1.key 1
mv n2.thresholdkey n2.key 2
mv n3.thresholdkey n3.key 3
mv n4.thresholdkey n4.key 4

cp *.pub *.crt 1
cp *.pub *.crt 2
cp *.pub *.crt 3
cp *.pub *.crt 4

rm *.pub *.crt

