#!/bin/bash

docker stop $(docker ps | grep 'n*' | awk '{print $1}')
docker container prune