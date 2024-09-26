# 在Docker容器中运行Slurm

容器共有一个head节点和8个node节点，在运行前需确保本机已经安装了MySQL或者以Docker运行了MySQL容器，如果没有，参考以下脚本安装MySQL

```bash
docker run -d --rm -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root --network=ohpc-container-network --name=mysql --hostname=mysql mysql
```

## 运行

```bash
./build.sh
./run.sh
```

