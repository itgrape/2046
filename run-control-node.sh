#!/bin/bash

echo "=== Start MySQL"
docker run -d --rm -e MYSQL_ROOT_PASSWORD=root --net=host --name=mysql --hostname=mysql mysql

echo "=== Start ipa-server"
docker run -d --rm --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup --net=host --name=ipa --hostname=ipa.example.com ohpc-container/ipa

echo "=== Start head node"
docker run -d --rm --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup --net=host --name=head --hostname=head.example.com ohpc-container/head


# test
# docker exec -it head /bin/bash
