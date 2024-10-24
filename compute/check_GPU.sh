#!/bin/bash

# 自定义值
GPU_THRESHOLD=20     # 设置GPU平均使用率阈值百分比，低于此值将释放作业
CHECK_INTERVAL=60    # GPU监控间隔时间（单位：秒）
MAX_HISTORY_SIZE=12  # 保存最近多少次的GPU使用率

exec >> /var/log/slurm/check_GPU.log 2>&1

JOB_ID=$SLURM_JOB_ID
USER_ID=$SLURM_JOB_USER

echo "Prolog script started for job $JOB_ID by user $USER_ID"

# 检查作业是否分配了GPU，没有分配GPU直接退出该脚本
if [[ -z $SLURM_JOB_GPUS ]]; then
  echo "No GPU allocated for job $JOB_ID. Exiting."
  exit 0
fi

# 获取分配的GPU
ALLOCATED_GPUS=$(echo $SLURM_JOB_GPUS | tr ',' ' ')

# 保存最近MAX_HISTORY_SIZE次的GPU使用率
GPU_UTIL_HISTORY=()

# 计算最近的GPU使用率平均值
calculate_average_utilization() {
  local sum=0
  local count=${#GPU_UTIL_HISTORY[@]}
  
  for util in "${GPU_UTIL_HISTORY[@]}"; do
    sum=$(($sum + $util))
  done

  # 如果还没有足够的记录，返回当前已有的平均值，-gt是大于的意思
  if [ $count -gt 0 ]; then
    echo $(($sum / $count))
  else
    echo 0
  fi
}

# 持续监控GPU使用率
while true; do
  # 获取所有分配GPU的使用率
  GPU_UTILIZATION=0
  for GPU in $ALLOCATED_GPUS; do
    UTIL=$(nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader,nounits -i $GPU | xargs)
    GPU_UTILIZATION=$(($GPU_UTILIZATION + $UTIL))
  done

  # 计算当前的平均使用率
  AVG_GPU_UTILIZATION=$(($GPU_UTILIZATION / $(echo $ALLOCATED_GPUS | wc -w)))

  # 记录最近的GPU使用率
  GPU_UTIL_HISTORY+=($AVG_GPU_UTILIZATION)
  if [ ${#GPU_UTIL_HISTORY[@]} -gt $MAX_HISTORY_SIZE ]; then
    GPU_UTIL_HISTORY=("${GPU_UTIL_HISTORY[@]:1}")  # 切片操作移除最早的一次记录，保持记录的总数为MAX_HISTORY_SIZE
  fi

  # 如果记录不足MAX_HISTORY_SIZE次，则等待
  if [ ${#GPU_UTIL_HISTORY[@]} -lt $MAX_HISTORY_SIZE ]; then
    echo "Job $JOB_ID by user $USER_ID, not enough data for checking yet. Current GPU usage: $AVG_GPU_UTILIZATION%"
  else
    # 计算最近MAX_HISTORY_SIZE次的GPU使用率平均值
    RECENT_AVG_UTILIZATION=$(calculate_average_utilization)

    echo "Job $JOB_ID by user $USER_ID, Current GPU usage: $AVG_GPU_UTILIZATION%, Recent Average GPU: $RECENT_AVG_UTILIZATION%"

    # 检查最近MAX_HISTORY_SIZE次的平均使用率是否低于阈值
    if (( $RECENT_AVG_UTILIZATION < $GPU_THRESHOLD )); then
      echo "Recent average GPU usage is below $GPU_THRESHOLD% for job $JOB_ID. Releasing the job."
      scancel $JOB_ID
      exit 0
    fi
  fi

  sleep $CHECK_INTERVAL
done
