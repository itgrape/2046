#!/bin/bash


# =========== 向守护进程注册 ===========
CPU_CHECKS=30
MONITOR_INTERVAL=60
MONITOR_LOG_PATH="${SLURM_SUBMIT_DIR}/monitor-${SLURM_JOB_ID}.log"

/usr/local/bin/job_helper register $CPU_CHECKS $MONITOR_LOG_PATH
if [ $? -ne 0 ]; then
    exit 1
fi

# =========== 启动监控进程 ===========
./job_helper monitor $MONITOR_INTERVAL &
MONITOR_PID=$!
mkdir -p "${HOME}/.monitor"
echo $MONITOR_PID > "${HOME}/.monitor/monitor-${SLURM_JOB_ID}.pid"



# ============ Dropbear ============
exec >> "${SLURM_SUBMIT_DIR}/connect-${SLURM_JOB_ID}.log" 2>&1

LOGIN_NODE_ADDRESS="10.10.20.2"
JOB_ID=$SLURM_JOB_ID
LOGIN_NODE_ALIAS="Slurm-Login"
COMPUTE_NODE_ALIAS="Job-${JOB_ID}"
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

PORT=$(generate_random_port)
NODE_HOSTNAME=$(hostname)
PID_FILE="${HOME}/.dropbear/dropbear-${SLURM_JOB_ID}.pid"

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
echo "--- SSH CONFIGURATION FOR YOUR LOCAL MACHINE (~/.ssh/config) ---"
echo "Job is running on compute node: ${NODE_HOSTNAME} with user: ${USER}"
echo "--------------------------------------------------------------------------------"
echo "1. COPY the entire block below and PASTE it into your LOCAL ~/.ssh/config file."
echo "2. If a Host with the same name already exists, please update or remove it."
echo "--------------------------------------------------------------------------------"

# 使用 cat <<EOF 来打印多行文本块，并让shell替换变量
cat << EOF

# Block for HPC Login Node (Jump Host)
Host ${LOGIN_NODE_ALIAS}
    HostName ${LOGIN_NODE_ADDRESS}
    User ${USER}
    IdentityFile ~/.ssh/id_rsa

# Block for your GPU Job (Connect through the Login Node)
Host ${COMPUTE_NODE_ALIAS}
    HostName ${NODE_HOSTNAME}
    User ${USER}
    IdentityFile ~/.ssh/id_rsa
    Port ${PORT}
    ProxyJump ${LOGIN_NODE_ALIAS}
    # Keep the connection alive
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

exit 0