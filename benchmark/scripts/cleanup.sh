#!/bin/bash

# PURPOSE: delete key, measurement and configuration files

rm -rf keys
rm hotcertification.toml

docker container prune -f