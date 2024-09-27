# 在Docker容器中运行Slurm

容器共有1个head节点和8个node节点，启动服务后将自动创建1个MySQL容器用于Slurmdbd相关服务

## 运行

```bash
./build.sh
./run.sh
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