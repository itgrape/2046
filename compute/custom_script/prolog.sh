#!/bin/bash

if [ "${SLURM_UID}" = "" ]; then
    exit 0
fi
if [ "${SLURM_JOB_ID}" = "" ]; then
    exit 0
fi
if [ "${SLURM_JOB_USER}" = "" ]; then
    exit 0
fi


echo "Prolog script started at $(date) for User ${SLURM_JOB_USER} Job ${SLURM_JOB_ID}" >> /var/log/slurm/prolog.log


unit_name="check_service_${SLURM_JOB_USER}_${SLURM_JOB_ID}"
if [[ "$SLURM_JOB_PARTITION" == "Debug" ]]; then
    systemd-run --unit=$unit_name --slice=background --setenv=SLURM_JOB_STDOUT=$SLURM_JOB_STDOUT --setenv=SLURM_JOB_ID=$SLURM_JOB_ID --setenv=SLURM_JOB_USER=$SLURM_JOB_USER --setenv=SLURM_JOB_GPUS=$SLURM_JOB_GPUS --setenv=THRESHOLD=0 bash "/etc/slurm/check_GPU.sh"
else
    systemd-run --unit=$unit_name --slice=background --setenv=SLURM_JOB_STDOUT=$SLURM_JOB_STDOUT --setenv=SLURM_JOB_ID=$SLURM_JOB_ID --setenv=SLURM_JOB_USER=$SLURM_JOB_USER --setenv=SLURM_JOB_GPUS=$SLURM_JOB_GPUS --setenv=THRESHOLD=50 bash "/etc/slurm/check_GPU.sh"
fi


exit 0