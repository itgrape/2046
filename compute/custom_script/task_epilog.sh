#!/bin/bash


# =========== 清理监控进程 ===========
PID_FILE="${HOME}/.monitor/monitor-${SLURM_JOB_ID}.pid"
if [ -f "$PID_FILE" ]; then
    MONITOR_PID=$(cat "$PID_FILE")
    if [ -n "$MONITOR_PID" ]; then
        if ps -p "$MONITOR_PID" > /dev/null; then
            kill "$MONITOR_PID"
            sleep 2
            if ps -p "$MONITOR_PID" > /dev/null; then
                kill -9 "$MONITOR_PID"
            fi
        fi
    fi
    rm -f "$PID_FILE"
fi


# =========== 清理 Dropbear SSH Server ===========
PID_FILE="${HOME}/.dropbear/dropbear-${SLURM_JOB_ID}.pid"
if [ -f "$PID_FILE" ]; then
    DROPBEAR_PID=$(cat "$PID_FILE")
    if [ -n "$DROPBEAR_PID" ]; then
        if ps -p "$DROPBEAR_PID" > /dev/null; then
            kill "$DROPBEAR_PID"
            sleep 2
            if ps -p "$DROPBEAR_PID" > /dev/null; then
                kill -9 "$DROPBEAR_PID"
            fi
        fi
    fi
    rm -f "$PID_FILE"
fi

exit 0
