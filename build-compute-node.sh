#!/bin/bash

CONTAINER=podman

echo "=== build compute"
$CONTAINER build -t ohpc-container/compute -f compute/Containerfile compute


$CONTAINER image prune -f