#!/bin/bash

cat /tmp/auth_files/passwd.bak  >> /etc/passwd
cat /tmp/auth_files/group.bak  >> /etc/group
cat /tmp/auth_files/gshadow.bak  >> /etc/gshadow
cat /tmp/auth_files/shadow.bak  >> /etc/shadow
\cp -rf /tmp/auth_files/login.defs /etc/login.defs

salt '*' state.apply sync_auth_files
