#!/bin/bash

input_file="/auth_files/user_info.txt"

if [ ! -f "$input_file" ]; then
    echo "File $input_file not found!"
    exit 1
fi

while read -r line; do
    if [[ "$line" =~ ^Username: ]]; then

        # read info
        username=$(echo "$line" | cut -d: -f2 | tr -d '[:space:]')
        read -r line
        password=$(echo "$line" | cut -d: -f2 | tr -d '[:space:]')
        read -r line
        group=$(echo "$line" | cut -d: -f2 | tr -d '[:space:]')
        read -r line
        uid=$(echo "$line" | cut -d: -f2 | tr -d '[:space:]')
        read -r line
        gid=$(echo "$line" | cut -d: -f2 | tr -d '[:space:]')
        read -r line
        home=$(echo "$line" | cut -d: -f2 | tr -d '[:space:]')
        read -r line
        shell=$(echo "$line" | cut -d: -f2 | tr -d '[:space:]')

        if id "$username" &>/dev/null; then
            echo "User $username already exists. Skipping."
        else
            # create group
            if ! getent group "$group" >/dev/null 2>&1; then
                groupadd "$group"
                echo "Group $group created."
            fi

            # create user
            useradd -m -u "$uid" -g "$gid" -d "$home" -s "$shell" "$username"
            echo "User $username created successfully."

            # set password
            usermod --password "$password" "$username"
            echo "Password for $username set."

            # add user to group
            usermod -aG "$group" "$username"
            echo "User $username added to group $group."
        fi
    fi
done < "$input_file"
