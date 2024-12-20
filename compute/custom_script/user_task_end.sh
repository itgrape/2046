#!/bin/bash
#
# This script will kill any user processes on a node when the last
# SLURM job there ends. For example, if a user directly logs into
# an allocated node SLURM will not kill that process without this
# script being executed as an epilog.
#
# SLURM_BIN can be used for testing with private version of SLURM
# SLURM_BIN="/usr/bin/"
#
if [ "${SLURM_UID}" = "" ]; then
    exit 0
fi
if [ "${SLURM_JOB_ID}" = "" ]; then
    exit 0
fi
if [ "${SLURM_JOB_USER}" = "" ]; then
    exit 0
fi
#
# Stop check GPU
#
unit_name="check_service_${SLURM_JOB_USER}_${SLURM_JOB_ID}"
systemctl stop $unit_name
echo "Stopped systemd service $unit_name for job $SLURM_JOB_ID" >> /var/log/slurm/epilog.log
#
# Don't try to kill user root or system daemon jobs
#
if [ "${SLURM_UID}" -lt 1000 ]; then
    exit 0
fi

job_list=$("${SLURM_BIN}"squeue --noheader --format=%A --user="${SLURM_UID}" --node=localhost)
for job_id in ${job_list}; do
if [ "${job_id}" -ne "${SLURM_JOB_ID}" ]; then
        exit 0
    fi
done

#
# No other SLURM jobs, purge all remaining processes of this user if and only
# if the SLURM controller is not running on the same server as the running job,
# to avoid killing the login session
#
if ! pgrep -x slurmctld >/dev/null; then
    pkill -KILL -U "${SLURM_UID}"
fi

exit 0
