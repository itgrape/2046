#!/bin/bash

CONTAINER=podman

# Which node
read -p "Please enter compute node order: " order

echo "=== Start compute-node-$order"
$CONTAINER run -d --rm \
    --add-host=head:10.20.20.1 \
    --add-host=compute-5:10.20.11.80 \
    --add-host=compute-4:10.20.11.81 \
    --add-host=compute-3:10.20.11.82 \
    --add-host=compute-2:10.20.11.83 \
    --add-host=compute-1:10.20.11.84 \
    --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
    --network=host \
    --name=compute-$order --hostname=compute-$order \
    --device nvidia.com/gpu=all \
    ohpc-container/compute

# test
$CONTAINER exec -it compute-$order /bin/bash
