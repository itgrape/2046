#!/bin/bash

output_file="/auth_files/user_info.txt"

if [ ! -f "$output_file" ]; then
    touch "$output_file"
fi


temp_file="/tmp/temp_user_info.txt"
> "$temp_file"


while IFS=: read -r username password uid gid gecos home shell; do
    if [ "$uid" -ge 1000 ] && [ "$username" != "nobody" ]; then
        group=$(getent group "$gid" | cut -d: -f1)
        shadow_entry=$(getent shadow "$username")
        user_password=$(echo "$shadow_entry" | cut -d: -f2)

        # 检查输出文件中是否已有该用户名的记录
        if grep -q "^Username: $username" "$output_file"; then
            # 如果用户信息已存在，则更新该用户的相关信息
            sed -i "/^Username: $username/,+7d" "$output_file"
        fi

        echo "Username: $username" >> "$temp_file"
        echo "Password: $user_password" >> "$temp_file"
        echo "Group: $group" >> "$temp_file"
        echo "UID: $uid" >> "$temp_file"
        echo "GID: $gid" >> "$temp_file"
        echo "Home Directory: $home" >> "$temp_file"
        echo "Shell: $shell" >> "$temp_file"
        echo "----------------------------------" >> "$temp_file"
    fi
done < /etc/passwd


cat "$temp_file" >> "$output_file"
rm -f "$temp_file"


echo "User information exported to $output_file"
