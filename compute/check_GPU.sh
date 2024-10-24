#!/bin/bash

exec >> /var/log/slurm/check_GPU.log 2>&1

JOB_ID=$SLURM_JOB_ID
USER_ID=$SLURM_JOB_USER
GPU_THRESHOLD=20  # 设置GPU平均使用率阈值百分比，低于此值将释放作业

# GPU监控间隔时间（单位：秒）
CHECK_INTERVAL=60

echo "Prolog script started for job $JOB_ID by user $USER_ID"

# 检查作业是否分配了GPU，没有启动GPU直接退出该脚本
if [[ -z $SLURM_JOB_GPUS ]]; then
  echo "No GPU allocated for this job. Exiting."
  exit 0
fi

# 申请资源后先睡一会，不然直接退出了
sleep 120
# 获取分配的 GPU
ALLOCATED_GPUS=$(echo $SLURM_JOB_GPUS | tr ',' ' ')

# 持续监控GPU使用率
while true; do
  # 获取所有分配GPU的使用率
  GPU_UTILIZATION=0
  for GPU in $ALLOCATED_GPUS; do
    UTIL=$(nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader,nounits -i $GPU | xargs)
    GPU_UTILIZATION=$(($GPU_UTILIZATION + $UTIL))
  done

  # 计算平均使用率
  AVG_GPU_UTILIZATION=$(($GPU_UTILIZATION / $(echo $ALLOCATED_GPUS | wc -w)))
  echo "Job $JOB_ID by user $USER_ID, GPU: $AVG_GPU_UTILIZATION"

  # 检查是否低于阈值
  if (( $AVG_GPU_UTILIZATION < $GPU_THRESHOLD )); then
    echo "GPU usage is below $GPU_THRESHOLD% for job $JOB_ID. Releasing the job."
    scancel $JOB_ID
    exit 0
  fi

  sleep $CHECK_INTERVAL
done
