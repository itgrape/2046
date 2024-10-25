#!/bin/bash

ssh-keygen -t rsa

HOSTS=("compute-0.example.com" "compute-1.example.com" "compute-2.example.com")
PORT=2222

for HOST in "${HOSTS[@]}"; do
  ssh-copy-id -p ${PORT} root@${HOST}
done
