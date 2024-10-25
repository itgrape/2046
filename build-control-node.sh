#!/bin/bash

# user=${USER:-$USER}
# arch=$(uname -m)

# set -e
# echo "=== build openhpc"
# docker build -t ohpc-container/openhpc:3 -f openhpc/Containerfile openhpc \
#   --build-arg USER=$user \
#   --build-arg ARCH=$arch

echo "=== build head"
docker build -t ohpc-container/head -f head/Containerfile head


docker image prune -f