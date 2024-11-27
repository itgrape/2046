#!/bin/bash

output_file="/auth_files/user_info.txt"

# clear the file
> "$output_file"

while IFS=: read -r username password uid gid gecos home shell; do
    if [ "$uid" -ge 1000 ] && [ "$username" != "nobody" ]; then
        group=$(getent group "$gid" | cut -d: -f1)
        shadow_entry=$(getent shadow "$username")
        user_password=$(echo "$shadow_entry" | cut -d: -f2)

        echo "Username: $username" >> "$output_file"
        echo "Password: $user_password" >> "$output_file"
        echo "Group: $group" >> "$output_file"
        echo "UID: $uid" >> "$output_file"
        echo "GID: $gid" >> "$output_file"
        echo "Home Directory: $home" >> "$output_file"
        echo "Shell: $shell" >> "$output_file"
        echo "----------------------------------" >> "$output_file"
    fi
done < /etc/passwd

echo "User information exported to $output_file"
