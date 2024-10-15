# 在Docker容器中运行Slurm

安装docker：https://docs.docker.com/engine/install/rhel/

容器共有一个head节点和数个compute节点，启动服务后将自动创建1个MySQL容器用于Slurmdbd相关服务

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

## 打包运行

```bash
# head
./build-control-node.sh
./run-control-node.sh
docker exec -it head /usr/tmp/ipa-server-install.sh

# compute
./build-compute-node.sh
./run-compute-node.sh
docker exec -it compute-x /usr/tmp/ipa-client-install.sh
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

echo "Test job running"
sleep 30
echo "Test job finished"
```

## 常用命令

ipa用户管理相关（管理员）

```bash
#初始化
kinit admin

#添加新用户
ipa user-add <username> --first=<first> --last=<last> --password

#修改用户密码
ipa passwd <username>

#删除用户
ipa user-del <username>
```

sacctmgr资源限制相关（管理员）

```bash
#添加cluster
sacctmgr add cluster <clustername>

#添加account
sacctmgr add account <accountname>

#添加user
sacctmgr add user <username> account=<accountname>

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

