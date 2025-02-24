#!/bin/bash

echo "Prolog script started at $(date) for User ${SLURM_JOB_USER} Job ${SLURM_JOB_ID}" >> /var/log/slurm/prolog.log


# 通过互斥锁确保同一时间文件只被一个进程访问
{
    flock -x 200

    JOB_INFO=$(squeue -u "$SLURM_JOB_USER" -o "%P" -h)
    if echo "$JOB_INFO" | grep -q "Debug" || [[ "$SLURM_JOB_PARTITION" == "Debug" ]]; then
        # 当前用户有 Debug 分区的作业
        sed -i "\|${SLURM_JOB_USER}|d" /etc/ssh/denied_users
    else
        echo "$SLURM_JOB_USER" >> /etc/ssh/denied_users
    fi
} 200>/tmp/denied_users.lock


unit_name="check_service_${SLURM_JOB_USER}_${SLURM_JOB_ID}"
if [[ "$SLURM_JOB_PARTITION" == "Debug" ]]; then
    systemd-run --unit=$unit_name --slice=background --setenv=SLURM_JOB_ID=$SLURM_JOB_ID --setenv=SLURM_JOB_USER=$SLURM_JOB_USER --setenv=SLURM_JOB_GPUS=$SLURM_JOB_GPUS --setenv=THRESHOLD=20 bash "/etc/slurm/check_GPU.sh"
else
    systemd-run --unit=$unit_name --slice=background --setenv=SLURM_JOB_ID=$SLURM_JOB_ID --setenv=SLURM_JOB_USER=$SLURM_JOB_USER --setenv=SLURM_JOB_GPUS=$SLURM_JOB_GPUS --setenv=THRESHOLD=50 bash "/etc/slurm/check_GPU.sh"
fi


exit 0