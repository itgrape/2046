#!/bin/bash

NODE_HOSTNAME=$(hostname)

# 设置锁文件，防止程序多次执行
LOCK_DIR="/tmp/slurm_locks_${USER}"
mkdir -p "$LOCK_DIR"
PROLOG_LOCK_FILE="${LOCK_DIR}/prolog-${SLURM_JOB_ID}-${NODE_HOSTNAME}.lock"

(
    flock -n 200 || { echo "[Prolog on ${NODE_HOSTNAME}]: Another instance is already running. Exiting."; exit 0; }

    # =======================================
    # ============= 监控程序系统 ==============
    # =======================================
    MONITOR_DIR="${HOME}/.monitor"
    mkdir -p "$MONITOR_DIR"
    (
        HELPER_PATH="/usr/local/bin/job_helper"
        INFO_LOG_PATH="${SLURM_SUBMIT_DIR}/info-${SLURM_JOB_ID}.log"

        # --- 注册任务信息 ---
        $HELPER_PATH register $INFO_LOG_PATH
        if [ $? -ne 0 ]; then
            echo "[Prolog on ${NODE_HOSTNAME}] Error: Job registration with monitoring daemon failed. Aborting."
            exit 1
        fi
        echo "[Prolog on ${NODE_HOSTNAME}] Registration successful."

        # --- 监控进程交给 at 管理 ---
        AT_JOB_ID=$(echo "$HELPER_PATH monitor" | at now 2>&1 | grep "job" | awk '{print $2}' | tail -n 1)
        echo "[Prolog on ${NODE_HOSTNAME}] Monitor process scheduled with 'at' job ID: $AT_JOB_ID."

    ) >> "${HOME}/.monitor/monitor-${SLURM_JOB_ID}-${NODE_HOSTNAME}.log" 2>&1




    # =======================================
    # =========== Dropbear 服务器 ============
    # =======================================
    (
        generate_random_port() {
            local port
            while true; do
                port=$((RANDOM % 10001 + 50000))  # 生成 50000 到 60000 之间的随机端口
                if ! netstat -tuln | grep -q ":$port "; then
                    echo "$port"
                    return
                fi
            done
        }

        LOGIN_NODE_ADDRESS="10.10.20.2"
        JOB_ID=$SLURM_JOB_ID
        LOGIN_NODE_ALIAS="Slurm-Login"

        COMPUTE_NODE_ALIAS="Job-${JOB_ID}-${NODE_HOSTNAME}"

        PORT=$(generate_random_port)
        PID_FILE="${HOME}/.dropbear/dropbear-${SLURM_JOB_ID}-${NODE_HOSTNAME}.pid"


        # --- Dropbear SSH Server Setup ---
        KEY_DIR="$HOME/.dropbear"
        mkdir -p "$KEY_DIR"
        HOST_KEY="$KEY_DIR/dropbear_rsa_host_key"
        if [ ! -f "$HOST_KEY" ]; then
            echo "Host key not found. Generating a new one at $HOST_KEY"
            dropbearkey -t rsa -f "$HOST_KEY"
        fi

        dropbear -r "$HOST_KEY" -p "$PORT" -P "$PID_FILE" -w -s


        echo "================================================================================"
        echo "--- SSH CONFIGURATION FOR NODE: ${NODE_HOSTNAME} ---"
        echo "--- User: ${USER} | Job ID: ${JOB_ID} ---"
        echo "--------------------------------------------------------------------------------"
        echo "1. COPY the entire block below and PASTE it into your LOCAL ~/.ssh/config file."
        echo "2. If a Host with the same name already exists, please update or remove it."
        echo "--------------------------------------------------------------------------------"

        cat << EOF

# Block for HPC Login Node (Jump Host)
Host ${LOGIN_NODE_ALIAS}
    HostName ${LOGIN_NODE_ADDRESS}
    User ${USER}
    IdentityFile ~/.ssh/id_rsa

# Block for your Job (Connect through the Login Node)
Host ${COMPUTE_NODE_ALIAS}
    HostName ${NODE_HOSTNAME}
    User ${USER}
    IdentityFile ~/.ssh/id_rsa
    Port ${PORT}
    ProxyJump ${LOGIN_NODE_ALIAS}
    ServerAliveInterval 60

EOF
        echo "================================================================================"
        echo "--- HOW TO CONNECT ---"
        echo ""
        echo ">>> For Command Line (Terminal):"
        echo "    ssh ${COMPUTE_NODE_ALIAS}"
        echo ""
        echo ">>> For VS Code (Remote-SSH Extension):"
        echo "    1. Click the green '><' icon in the bottom-left corner."
        echo "    2. Select 'Connect to Host...' or 'Connect Current Window to Host...'."
        echo "    3. Choose '${COMPUTE_NODE_ALIAS}' from the list."
        echo ""
        echo "================================================================================"
    ) >> "${SLURM_SUBMIT_DIR}/connect-${SLURM_JOB_ID}.log" 2>&1

) 200>"$PROLOG_LOCK_FILE"
# flock 会将文件描述符 200 指向锁文件，在括号内的命令执行期间保持锁定
# 当命令结束时，文件描述符关闭，锁被自动释放



exit 0