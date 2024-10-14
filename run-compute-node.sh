#!/bin/bash

# Which node
read -p "Please enter compute node order: " order


echo "=== Start compute-node-$order"
docker run -d --rm --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup --net=host --name=compute-$order --hostname=compute-$order.example.com ohpc-container/node


# test
# docker exec -it compute-0 /bin/bash
