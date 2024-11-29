#!/bin/bash

# 用户备份文件路径
output_file="/auth_files/user_info.txt"

# 如果文件不存在则创建文件
if [ ! -f "$output_file" ]; then
    touch "$output_file"
fi

# 删除用户并同步更新备份文件
delete_user() {
    local username=$1
    local remove_home=$2

    # 检查用户是否存在
    if ! id "$username" &>/dev/null; then
        echo "User $username does not exist."
        return 1
    fi

    # 删除系统用户，并根据是否有 -r 参数删除用户主目录
    if [ "$remove_home" == "-r" ]; then
        userdel -r "$username"
    else
        userdel "$username"
    fi

    if [ $? -ne 0 ]; then
        echo "Failed to delete user $username."
        return 1
    fi

    # 从备份文件中删除该用户的信息
    if grep -q "^Username: $username" "$output_file"; then
        sed -i "/^Username: $username/,+7d" "$output_file"
        echo "User $username information deleted from backup file."
    else
        echo "No information found for user $username in the backup file."
    fi
}

# 主操作
if [ -z "$1" ]; then
    echo "Usage: $0 <username> [-r]"
    exit 1
fi

# 获取用户名
username=$1

# 检查是否使用了 -r 参数
if [ "$2" == "-r" ]; then
    delete_user "$username" "-r"
else
    delete_user "$username"
fi
