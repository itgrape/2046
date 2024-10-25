#!/bin/bash


SOURCE_FILES=("/etc/passwd" "/etc/shadow" "/etc/group" "/etc/gshadow" "/etc/login.defs")
DEST_PATH="/etc/"
HOSTS=("compute-0.example.com" "compute-1.example.com" "compute-2.example.com")
PORT=2222

for HOST in "${HOSTS[@]}"; do
  for FILE in "${SOURCE_FILES[@]}"; do
    echo "正在将文件 ${FILE} 复制到 root@${HOST}:${DEST_PATH} ..."
    scp "-P ${PORT}" "$FILE" "root@${HOST}:${DEST_PATH}"
    
    # 检查是否复制成功
    if [ $? -eq 0 ]; then
      echo "文件 ${FILE} 成功复制到 ${HOST}"
    else
      echo "文件 ${FILE} 复制到 ${HOST} 失败"
    fi
  done
done
