#!/bin/bash

echo "=== build compute"
docker build -t ohpc-container/compute -f compute/Containerfile compute


docker image prune -f