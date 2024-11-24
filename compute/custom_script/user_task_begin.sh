#!/bin/bash

echo "Prolog script started at $(date) for User ${SLURM_JOB_USER} Job ${SLURM_JOB_ID}" >> /var/log/slurm/prolog.log

unit_name="check_service_${SLURM_JOB_USER}_${SLURM_JOB_ID}"

systemd-run --unit=$unit_name --slice=background --setenv=SLURM_JOB_ID=$SLURM_JOB_ID --setenv=SLURM_JOB_USER=$SLURM_JOB_USER --setenv=SLURM_JOB_GPUS=$SLURM_JOB_GPUS bash "/etc/slurm/check_GPU.sh"

exit 0