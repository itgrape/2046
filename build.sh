#!/bin/bash

user=${USER:-$USER}
arch=$(uname -m)

# echo "=== setup docker"
# docker network create ohpc-container-network

set -e
echo "=== build openhpc"
docker build -t ohpc-container/openhpc:3 -f openhpc/Containerfile openhpc \
  --build-arg USER=$user \
  --build-arg ARCH=$arch

echo "=== build head"
docker build -t ohpc-container/head -f head/Containerfile head

echo "=== build node"
docker build -t ohpc-container/node -f node/Containerfile node


docker image prune -f