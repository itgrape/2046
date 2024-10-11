#!/bin/bash

# user=${USER:-$USER}
# arch=$(uname -m)

# # echo "=== setup docker"
# # docker network create ohpc-container-network

# set -e
# echo '=== build openhpc'
# docker build -t ohpc-container/openhpc3:systemd -f openhpc/Containerfile openhpc \
#   --build-arg USER=$user \
#   --build-arg ARCH=$arch

for I in head node ; do
  echo "=== build $I"
  docker build -t ohpc-container/$I:systemd -f $I/Containerfile $I
done

docker image prune -f