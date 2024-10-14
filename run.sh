#!/bin/bash

# volume=--volume=/home:/home
network=--network=ohpc-container-network

# echo "=== Start MySQL"
# docker run -d --rm -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root $network --name=mysql --hostname=mysql mysql

echo "=== Start head"
docker run -d --rm --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup -p 2222:22 $network --name=head --hostname=head.example.com ohpc-container/head

# echo "=== Start ipa-server"
# docker run -d --rm --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup $network --name=ipa --hostname=ipa.example.com ohpc-container/ipa

# echo "=== Start node"
# docker run -d --rm --privileged --cgroupns=host -v /sys/fs/cgroup:/sys/fs/cgroup $network --name=node-0 --hostname=node-0 ohpc-container/node


# test
# docker exec -it head /bin/bash
