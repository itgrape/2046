#!/bin/bash

exec >> /var/log/slurm/check_CPU.log 2>&1

CPU_UTILIZATION_THRESHOLD=50           # 设置CPU平均使用率阈值百分比，低于此值将释放作业
MEM_UTILIZATION_THRESHOLD=50           # 设置内存平均使用率阈值百分比，低于此值将释放作业
CHECK_INTERVAL=60                      # GPU监控间隔时间（单位：秒）
MAX_HISTORY_SIZE=30                    # 求最近多少次监控的平均值


# 确保必要的Slurm环境变量已设置
if [ -z "$SLURM_JOB_ID" ] || [ -z "$SLURM_JOB_USER" ]; then
  echo "$(date): Error: environment (SLURM_JOB_ID, SLURM_JOB_USER) is empty. exiting..."
  exit 1
fi

JOB_ID=$SLURM_JOB_ID
USER_ID=$SLURM_JOB_USER
JOB_STDOUT=$SLURM_JOB_STDOUT

echo "Check CPU and memory script started for job $JOB_ID by user $USER_ID"


CPU_UTIL_HISTORY=()
MEM_UTIL_HISTORY=()


calculate_average() {
  local -n arr=$1  # 使用nameref传递数组以避免复制
  local sum=0
  local count=${#arr[@]}

  if [ $count -eq 0 ]; then
    echo 0
    return
  fi

  for val in "${arr[@]}"; do
    sum=$((sum + val))
  done

  echo $((sum / count))
}


# 解析申请的总内存大小
MEM_TOTAL_MB_RAW=$(scontrol show job $JOB_ID | grep -o 'mem=[^,]*' | cut -d'=' -f2)
if [[ $MEM_TOTAL_MB_RAW == *G ]]; then
    # 从字符串中移除 'G' 并乘以 1024
    MEM_TOTAL_MB=$(echo "$MEM_TOTAL_MB_RAW" | sed 's/G//' | awk '{printf "%.0f", $1 * 1024}')
elif [[ $MEM_TOTAL_MB_RAW == *M ]]; then
    # 从字符串中移除 'M'
    MEM_TOTAL_MB=$(echo "$MEM_TOTAL_MB_RAW" | sed 's/M//' | awk '{printf "%.0f", $1}')
else
    # 如果没有单位，则假定单位是 MB
    MEM_TOTAL_MB=$(awk -v val="$MEM_TOTAL_MB_RAW" 'BEGIN {printf "%.0f", val}')
fi


while true; do
  CPU_UTILIZATION=$(ps -u $USER_ID -o pcpu --no-headers | awk '{sum+=$1} END {print int(sum)}')
  
  MEM_USED_MB=$(ps -u $USER_ID -o rss --no-headers | awk '{sum+=$1} END {print int(sum/1024)}')
  MEM_UTILIZATION=0
  if [ "$MEM_TOTAL_MB" -gt 0 ]; then
      MEM_UTILIZATION=$(echo "scale=0; ($MEM_USED_MB * 100) / $MEM_TOTAL_MB" | bc)
  fi

  CPU_UTIL_HISTORY+=($CPU_UTILIZATION)
  MEM_UTIL_HISTORY+=($MEM_UTILIZATION)
  if [ ${#CPU_UTIL_HISTORY[@]} -gt $MAX_HISTORY_SIZE ]; then
    CPU_UTIL_HISTORY=("${CPU_UTIL_HISTORY[@]:1}")
  fi
  if [ ${#MEM_UTIL_HISTORY[@]} -gt $MAX_HISTORY_SIZE ]; then
    MEM_UTIL_HISTORY=("${MEM_UTIL_HISTORY[@]:1}")
  fi

  # 如果记录不足MAX_HISTORY_SIZE次，则等待
  if [ ${#CPU_UTIL_HISTORY[@]} -lt $MAX_HISTORY_SIZE ]; then
    echo "Job $JOB_ID by user $USER_ID, not enough data for checking yet. Current CPU usage: $CPU_UTILIZATION%, Current memory usage: $MEM_UTILIZATION%"
  else

    RECENT_AVG_CPU_UTILIZATION=$(calculate_average CPU_UTIL_HISTORY)
    RECENT_AVG_MEM_UTILIZATION=$(calculate_average MEM_UTIL_HISTORY)

    echo "Job $JOB_ID by user $USER_ID, Current CPU usage: $CPU_UTILIZATION%, Recent Average CPU usage: $RECENT_AVG_CPU_UTILIZATION% , Current memory usage: $MEM_UTILIZATION%, Recent Average memory usage: $RECENT_AVG_MEM_UTILIZATION%"

    # 检查最近MAX_HISTORY_SIZE次的平均使用率是否低于阈值
    if (( $RECENT_AVG_CPU_UTILIZATION < $CPU_UTILIZATION_THRESHOLD )) || (( $RECENT_AVG_MEM_UTILIZATION < $MEM_UTILIZATION_THRESHOLD )); then
      echo "Recent average CPU usage or memory usage is too low for job $JOB_ID. Releasing the job."
      
      # 给用户发送为什么释放资源
      echo -e "\n--- SLURM JOB CANCELLED BY MONITOR ---" >> "$JOB_STDOUT"
      echo "Job $JOB_ID by user $USER_ID, Current CPU usage: $CPU_UTILIZATION%, Recent Average CPU usage: $RECENT_AVG_CPU_UTILIZATION% , Current memory usage: $MEM_UTILIZATION%, Recent Average memory usage: $RECENT_AVG_MEM_UTILIZATION%. Recent average CPU usage or memory usage is too low for job $JOB_ID. Releasing the job." >> $JOB_STDOUT
      echo "---------------------------------------" >> "$JOB_STDOUT"

      scancel $JOB_ID
      exit 0
    fi
  fi

  sleep $CHECK_INTERVAL
done
