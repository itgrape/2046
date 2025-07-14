#!/bin/bash

NODE_HOSTNAME=$(hostname)

echo "[Epilog on ${NODE_HOSTNAME}] Cleaning up job processes..."


# =======================================
# ============= 清理监控进程 ==============
# =======================================
MONITOR_AT_JOB_FILE="${HOME}/.monitor/monitor-at-jobid-${SLURM_JOB_ID}-${NODE_HOSTNAME}.txt"

if [ -f "$MONITOR_AT_JOB_FILE" ]; then
    AT_JOB_ID=$(cat "$MONITOR_AT_JOB_FILE")
    if [ -n "$AT_JOB_ID" ]; then
        echo "[Epilog on ${NODE_HOSTNAME}] Deleting 'at' job with ID ${AT_JOB_ID}."
        atrm "$AT_JOB_ID" 2>/dev/null
    fi
    rm -f "$MONITOR_AT_JOB_FILE"
else
    echo "[Epilog on ${NODE_HOSTNAME}] Monitor 'at' job ID file not found. Nothing to clean for monitor."
fi


# ========================================
# ========== 清理 Dropbear 进程 ===========
# ========================================
DROPBEAR_PID_FILE="${HOME}/.dropbear/dropbear-${SLURM_JOB_ID}-${NODE_HOSTNAME}.pid"

if [ -f "$DROPBEAR_PID_FILE" ]; then
    DROPBEAR_PID=$(cat "$DROPBEAR_PID_FILE")
    if [ -n "$DROPBEAR_PID" ] && ps -p "$DROPBEAR_PID" > /dev/null; then
        echo "[Epilog on ${NODE_HOSTNAME}] Stopping Dropbear process with PID ${DROPBEAR_PID}."
        kill "$DROPBEAR_PID"
        sleep 2
        # Force kill if it's still running
        if ps -p "$DROPBEAR_PID" > /dev/null; then
            kill -9 "$DROPBEAR_PID"
        fi
    fi
    rm -f "$DROPBEAR_PID_FILE"
else
    echo "[Epilog on ${NODE_HOSTNAME}] Dropbear PID file not found. Nothing to clean."
fi

echo "[Epilog on ${NODE_HOSTNAME}] Cleanup finished."

exit 0
