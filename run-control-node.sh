#!/bin/bash

echo "=== Start MySQL"
docker run -d --rm \
    -e MYSQL_ROOT_PASSWORD=root \
    --net=host \
    --name=mysql --hostname=mysql \
    mysql


echo "=== Start head node"
docker run -d --rm \
    --add-host=head.example.com:10.10.110.201 \
    --add-host=compute-0.example.com:10.10.110.202 \
    --add-host=compute-1.example.com:10.10.110.203 \
    --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup \
    --net=host \
    --name=head --hostname=head.example.com \
    ohpc-container/head


# test
docker exec -it head /bin/bash
