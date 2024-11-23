# 在容器中运行Slurm集群

安装docker：https://docs.docker.com/engine/install/rhel/

安装podman：https://podman.io/docs/installation

## 介绍

集群共有一个 control(控制) 节点 和数个 compute(计算) 节点，相关容器有

- MySQL容器，数据存储
- Head容器，Slurm控制节点
- Compute容器，Slurm计算节点

## 制作镜像

### 一、修改配置文件

1. run-control-node.sh 和 run-compute-node.sh 中按需增删 `--add-host=hostname:ip \` 用于配置ip与主机名的映射
2. head/gres.conf 中按需修改 GPU 参数信息
3. head/slurm.conf 中按需修改以下配置项
   - NodeName
   - PartitionName
4. compute/check_GPU.sh 中 5~8 行按需修改监控信息
5. compute/slurmd_override.conf 中按需修改 `ExecStart=/usr/sbin/slurmd --conf-server head-hostname:6817` 用于计算节点的无配置模式

### 二、制作镜像

运行如下命令

```
sudo ./build-control-node.sh
sudo ./build-compute-node.sh
```

## 部署运行

### 一、开放主机端口

1. 在 control 节点执行如下命令

```bash
sudo firewall-cmd --zone=public --add-port=2377/tcp --permanent
sudo firewall-cmd --zone=public --add-port=7946/tcp --permanent
sudo firewall-cmd --zone=public --add-port=9100/tcp --permanent
sudo firewall-cmd --zone=public --add-port=2222/tcp --permanent
sudo firewall-cmd --zone=public --add-port=6819/tcp --permanent
sudo firewall-cmd --zone=public --add-port=6817/tcp --permanent
sudo firewall-cmd --zone=public --add-port=6818/tcp --permanent
sudo firewall-cmd --zone=public --add-port=3306/tcp --permanent
sudo firewall-cmd --zone=public --add-port=389/tcp --permanent
sudo firewall-cmd --zone=public --add-port=636/tcp --permanent
sudo firewall-cmd --zone=public --add-port=88/tcp --permanent
sudo firewall-cmd --zone=public --add-port=464/tcp --permanent
sudo firewall-cmd --zone=public --add-port=53/tcp --permanent
sudo firewall-cmd --zone=public --add-port=80/tcp --permanent
sudo firewall-cmd --zone=public --add-port=443/tcp --permanent
sudo firewall-cmd --zone=public --add-port=7389/tcp --permanent
sudo firewall-cmd --zone=public --add-port=9443/tcp --permanent
sudo firewall-cmd --zone=public --add-port=8080/tcp --permanent
sudo firewall-cmd --zone=public --add-port=8443/tcp --permanent
sudo firewall-cmd --zone=public --add-port=7946/udp --permanent
sudo firewall-cmd --zone=public --add-port=4789/udp --permanent
sudo firewall-cmd --zone=public --add-port=88/udp --permanent
sudo firewall-cmd --zone=public --add-port=464/udp --permanent
sudo firewall-cmd --zone=public --add-port=53/udp --permanent
sudo firewall-cmd --zone=public --add-port=123/udp --permanent
sudo firewall-cmd --reload
```

2. 在 compute 节点执行如下命令

```
sudo firewall-cmd --zone=public --add-port=2377/tcp --permanent
sudo firewall-cmd --zone=public --add-port=7946/tcp --permanent
sudo firewall-cmd --zone=public --add-port=6818/tcp --permanent
sudo firewall-cmd --zone=public --add-port=2222/tcp --permanent
sudo firewall-cmd --zone=public --add-port=7946/udp --permanent
sudo firewall-cmd --zone=public --add-port=4789/udp --permanent
sudo firewall-cmd --reload
```

### 二、配置podman访问GPU

> 以下操作只在 compute 节点执行

https://podman-desktop.io/docs/podman/gpu

podman并不像docker集成了对gpu的支持，如果想在podman容器中使用GPU，需要借助nvidia-container-toolkit，具体步骤如下（在宿主机执行）

1. 安装nvidia-container-toolkit

```bash
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.repo | sudo tee /etc/yum.repos.d/nvidia-container-toolkit.repo

sudo dnf clean expire-cache
sudo dnf install -y nvidia-container-toolkit
```

2. 生成CDI文件

```bash
nvidia-ctk cdi generate --output=/etc/cdi/nvidia.yaml
nvidia-ctk cdi list   # Check
```

### 三、部署

1. 运行 control 节点

```
sudo ./run-control-node.sh
```

2. run-compute-node.sh 中按需修改 `--device nvidia.com/gpu=all \` 用于控制容器对 GPU 的访问

3. 运行 compute 节点

```
sudo ./run-compute-node.sh
```

## 测试

### 一、测试任务脚本

```bash
#!/bin/bash
#SBATCH --job-name=test_job
#SBATCH --output=/var/tmp/test_job%j_output.log
#SBATCH --error=/var/tmp/test_job%j_error.log
#SBATCH --ntasks=1
#SBATCH --time=00:01:00
#SBATCH --partition=normal
#SBATCH --mail-user=your_email@example.com
#SBATCH --mail-type=BEGIN,END,FAIL

echo "Test job running"
sleep 30
echo "Test job finished"
```

### 二、测试调试任务

```
salloc
```

## 常用命令

sacctmgr资源限制相关（管理员）

```bash
#添加cluster
sacctmgr add cluster <clustername>

#添加account
sacctmgr add account <accountname>

#添加user
sacctmgr add user <username> account=<accountname>

#删除user
sacctmgr remove user <username>

#创建qos并进行资源限制
sacctmgr add qos <qosname> MaxJobs=1 ...

#关联qos到用户或者账户
sacctmgr modify user name=<username> set qos=<qosname>
```

