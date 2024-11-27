#!/bin/bash

exec >> /var/log/slurm/check_GPU.log 2>&1

GPU_UTILIZATION_THRESHOLD=20   # 设置GPU平均使用率阈值百分比，低于此值将释放作业
GPU_MEMORY_THRESHOLD=20        # 设置GPU平均内存使用率阈值百分比，低于此值将释放作业
CHECK_INTERVAL=60              # GPU监控间隔时间（单位：秒）
MAX_HISTORY_SIZE=20            # 求最近多少次监控的平均值

JOB_ID=$SLURM_JOB_ID
USER_ID=$SLURM_JOB_USER

if [[ -z $SLURM_JOB_GPUS ]]; then
  echo "No GPU allocated for job $JOB_ID. Exiting."
  exit 0
fi

echo "Check GPU script started for job $JOB_ID by user $USER_ID"


ALLOCATED_GPUS=$(echo $SLURM_JOB_GPUS | tr ',' ' ')

GPU_UTIL_HISTORY=()
GPU_MEM_HISTORY=()


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
calculate_average_memory() {
  local sum=0
  local count=${#GPU_MEM_HISTORY[@]}
  
  for mem in "${GPU_MEM_HISTORY[@]}"; do
    sum=$(($sum + $mem))
  done

  if [ $count -gt 0 ]; then
    echo $(($sum / $count))
  else
    echo 0
  fi
}


while true; do
  GPU_UTILIZATION=0
  GPU_MEMORY=0
  for GPU in $ALLOCATED_GPUS; do
    UTIL=$(nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader,nounits -i $GPU | xargs)
    GPU_UTILIZATION=$(($GPU_UTILIZATION + $UTIL))
    MEM=$(nvidia-smi --query-gpu=memory.used,memory.total --format=csv,noheader,nounits -i $GPU | awk '{print ($1/$2)*100}' | xargs)
    GPU_MEMORY=$(($GPU_MEMORY + ${MEM%.*}))
  done

  AVG_GPU_UTILIZATION=$(($GPU_UTILIZATION / $(echo $ALLOCATED_GPUS | wc -w)))
  AVG_GPU_MEMORY=$(($GPU_MEMORY / $(echo $ALLOCATED_GPUS | wc -w)))

  GPU_UTIL_HISTORY+=($AVG_GPU_UTILIZATION)
  GPU_MEM_HISTORY+=($AVG_GPU_MEMORY)
  if [ ${#GPU_UTIL_HISTORY[@]} -gt $MAX_HISTORY_SIZE ]; then
    GPU_UTIL_HISTORY=("${GPU_UTIL_HISTORY[@]:1}")
  fi
  if [ ${#GPU_MEM_HISTORY[@]} -gt $MAX_HISTORY_SIZE ]; then
    GPU_MEM_HISTORY=("${GPU_MEM_HISTORY[@]:1}")
  fi

  # 如果记录不足MAX_HISTORY_SIZE次，则等待
  if [ ${#GPU_UTIL_HISTORY[@]} -lt $MAX_HISTORY_SIZE ]; then
    echo "Job $JOB_ID by user $USER_ID, not enough data for checking yet. Current GPU usage: $AVG_GPU_UTILIZATION%, Current GPU memory usage: $AVG_GPU_MEMORY%"
  else

    RECENT_AVG_UTILIZATION=$(calculate_average_utilization)
    RECENT_AVG_MEMORY=$(calculate_average_memory)

    echo "Job $JOB_ID by user $USER_ID, Current GPU usage: $AVG_GPU_UTILIZATION%, Recent Average GPU usage: $RECENT_AVG_UTILIZATION% , Current GPU memory usage: $AVG_GPU_MEMORY%, Recent Average GPU memory usage: $RECENT_AVG_MEMORY%"

    # 检查最近MAX_HISTORY_SIZE次的平均使用率是否低于阈值
    if (( $RECENT_AVG_UTILIZATION < $GPU_UTILIZATION_THRESHOLD )) || (( $RECENT_AVG_MEMORY < $GPU_MEMORY_THRESHOLD )); then
      echo "Recent average GPU usage is too low for job $JOB_ID. Releasing the job."
      scancel $JOB_ID
      exit 0
    fi
  fi

  sleep $CHECK_INTERVAL
done
