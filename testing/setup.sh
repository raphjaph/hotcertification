#!/bin/bash

pos allocations allocate bitcoingold

pos nodes image bitcoingold debian-buster

pos nodes reset bitcoingold

scp ./hotcertification bitcoingold:~

pos commands launch bitcoingold -- sh docker/install-docker-debian.sh


