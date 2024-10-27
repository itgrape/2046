#!/bin/bash

echo "=== Start MySQL container"
docker run -d --rm \
    -e MYSQL_ROOT_PASSWORD=root \
    --net=host \
    --name=mysql --hostname=mysql \
    mysql


echo "=== Start head container"
docker run -d --rm \
    --add-host=head.example.com:10.10.110.201 \
    --add-host=compute-0.example.com:10.10.110.202 \
    --add-host=compute-1.example.com:10.10.110.203 \
    --add-host=compute-2.example.com:10.10.110.41 \
    --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
    --net=host \
    --name=head --hostname=head.example.com \
    ohpc-container/head


# test
docker exec -it head /bin/bash
