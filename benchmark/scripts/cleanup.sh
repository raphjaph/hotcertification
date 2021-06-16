#!/bin/bash

# PURPOSE: delete key, measurement and configuration files

rm -rf measurements
rm -rf keys

docker container prune -f