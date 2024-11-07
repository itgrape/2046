#!/bin/bash
exec >> /var/log/slurm/monitor_user.log 2>&1

FILE_TO_WATCH=("/etc/passwd" "/etc/shadow" "/etc/group" "/etc/gshadow" "/etc/login.defs")


inotifywait -m -e modify,open,access,close_write,attrib,move,create,delete,close "${FILE_TO_WATCH[@]}" | while read path action file; do
    echo "File $FILE_TO_WATCH modified, running script..."

    cp /etc/passwd /etc/shadow /etc/group /etc/gshadow /etc/login.defs /srv/salt/auth_files
    chmod 644 /srv/salt/auth_files/shadow /srv/salt/auth_files/gshadow
    salt '*' state.apply sync_auth_files
done
