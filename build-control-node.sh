#!/bin/bash

CONTAINER=podman
ARCH=$(uname -m)

set -e
echo "=== build openhpc"
$CONTAINER build -t ohpc-container/openhpc:3 -f openhpc/Containerfile openhpc \
  --build-arg ARCH=$ARCH

echo "=== build head"
$CONTAINER build -t ohpc-container/head -f head/Containerfile head


$CONTAINER image prune -f