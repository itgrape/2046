#!/bin/bash

network=--network=ohpc-container-network
volume=--volume=ohpc-container-project:/project

echo "=== Start cluster"
docker run -d --rm -p 2222:22 $network $volume --name=head --hostname=head ohpc-container/head
for I in {0..7} ; do
  docker run -d --rm $network $volume --name=node-$I --hostname=node-$I ohpc-container/node
done


# test
# docker run -d -p 2222:22 --rm --network=ohpc-container-network --name=head --hostname=head ohpc-container/head
# docker exec -it head /bin/bash