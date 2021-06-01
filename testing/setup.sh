#!/bin/bash

#pos allocations allocate bitcoingold

#pos nodes image bitcoingold debian-buster

#pos nodes reset bitcoingold

scp -r /home/schleith/hotcertification bitcoingold:~

pos commands launch bitcoingold -- sh hotcertification/testing/install-docker-debian.sh


