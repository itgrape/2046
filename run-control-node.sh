#!/bin/bash

CONTAINER=podman

echo "=== Start MySQL container"
$CONTAINER run -d --rm \
    -e MYSQL_ROOT_PASSWORD=root \
    --network=host \
    --name=mysql --hostname=mysql \
    -v /root/mysql_data:/var/lib/mysql \
    mysql


echo "=== Start head container"
$CONTAINER run -d --rm \
    --add-host=head:10.20.20.1 \
    --add-host=compute-5:10.20.11.80 \
    --add-host=compute-4:10.20.11.81 \
    --add-host=compute-3:10.20.11.82 \
    --add-host=compute-2:10.20.11.83 \
    --add-host=compute-1:10.20.11.84 \
    --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
    --network=host \
    --name=head --hostname=head \
    ohpc-container/head
$CONTAINER cp /root/auth_files head:/srv/salt/recover


# test
$CONTAINER exec -it head /bin/bash
