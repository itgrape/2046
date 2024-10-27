#!/bin/bash

CONTAINER=$(command -v podman || command -v docker)

echo "=== build compute"
$CONTAINER build -t ohpc-container/compute -f compute/Containerfile compute


$CONTAINER image prune -f