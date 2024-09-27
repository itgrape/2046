#!/bin/bash

echo "=== Kill cluster"
docker kill head
for I in {0..7} ; do
  docker kill node-$I
done