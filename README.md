# 在容器中运行Slurm集群

安装docker：https://docs.docker.com/engine/install/rhel/

安装podman：https://podman.io/docs/installation

集群共有一个control节点和数个compute节点，相关容器有：

- MySQL容器，数据存储
- Head容器，Slurm控制节点
- Compute容器，Slurm计算节点

## 开放主机端口

```bash
# head
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

# compute
sudo firewall-cmd --zone=public --add-port=2377/tcp --permanent
sudo firewall-cmd --zone=public --add-port=7946/tcp --permanent
sudo firewall-cmd --zone=public --add-port=6818/tcp --permanent
sudo firewall-cmd --zone=public --add-port=2222/tcp --permanent
sudo firewall-cmd --zone=public --add-port=7946/udp --permanent
sudo firewall-cmd --zone=public --add-port=4789/udp --permanent
sudo firewall-cmd --reload
```

## 配置Hosts

由于dockerfile中无法直接修改/etc/host文件，所以应在docker run命令时使用--add-host添加。需根据实际情况修改run-xxx-node.sh中的相应部分

## 配置podman访问GPU

https://podman-desktop.io/docs/podman/gpu

podman并不像docker集成了对gpu的支持，如果想在podman容器中使用GPU，需要借助nvidia-container-toolkit，具体步骤如下（在宿主机执行）

安装nvidia-container-toolkit

```bash
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.repo | sudo tee /etc/yum.repos.d/nvidia-container-toolkit.repo

sudo dnf clean expire-cache
sudo dnf install -y nvidia-container-toolkit
```

生成CDI文件

```bash
nvidia-ctk cdi generate --output=/etc/cdi/nvidia.yaml
nvidia-ctk cdi list   # Check
```

## 打包运行

```bash
# head
./build-control-node.sh
./run-control-node.sh

# compute
./build-compute-node.sh
./run-compute-node.sh
```

## 测试任务脚本

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

申请资源（普通用户）

```bash
#提交批处理任务
sbatch xxx.sh

#申请交互式窗口
salloc (--no-shell)
```

