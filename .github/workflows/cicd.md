# CI/CD 部署指南

## 1. GitHub Actions 密钥配置

在 GitHub 仓库的 Settings -> Secrets and variables -> Actions 中设置以下密钥：

| 密钥名 | 说明 | 获取方式 |
|--------|------|----------|
| DOCKERHUB_USERNAME | Docker Hub 用户名 | - |
| DOCKERHUB_TOKEN | Docker Hub 访问令牌 | https://app.docker.com/settings/personal-access-tokens |
| SSH_PRIVATE_KEY | 用于连接生产服务器的 SSH 私钥 | 见下方说明 |
| KNOWN_HOSTS | 生产服务器的 SSH 指纹 | 见下方说明 |
| DEPLOY_HOST | 生产服务器的 IP 地址或域名 | - |
| DEPLOY_USER | 生产服务器的 SSH 用户名 | - |

### SSH 密钥配置步骤

1. 在服务器上生成 SSH 密钥：
```bash
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"  # 建议直接回车，不设置密码
```

2. 设置正确的权限：
```bash
chmod 700 ~/.ssh
chmod 600 ~/.ssh/id_rsa
chmod 644 ~/.ssh/id_rsa.pub
```

3. 配置授权：
```bash
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

4. 获取密钥内容：
```bash
# 获取 SSH_PRIVATE_KEY
cat ~/.ssh/id_rsa  # 复制整个输出内容（包括开头和结尾的标记）

# 获取 KNOWN_HOSTS
ssh-keyscan -H localhost > ~/.ssh/known_hosts
cat ~/.ssh/known_hosts
```

## 2. 服务器连接指南

### 2.1 同一内网环境

1. 测试连通性：
```bash
ping 私有IP地址
ssh 用户名@私有IP地址
```

### 2.2 不同网络环境

选择以下方式之一：
1. VPN 连接到同一网络
2. 通过跳板机连接：
```bash
ssh -J 用户名@跳板机公网IP 用户名@目标服务器私有IP
```

### 2.3 云服务器环境

1. 登录云服务控制台
2. 找到对应的 ECS 实例
3. 查看实例的公网IP或EIP
4. 连接服务器：
```bash
ssh 用户名@公网IP
```

### 2.4 内网服务器查找

```bash
# 扫描内网活跃主机
nmap -sP 192.168.1.0/24

# 详细扫描特定主机
nmap -A 私有IP地址

# 测试连接
ping 私有IP地址
ssh 用户名@私有IP地址
```

## 3. 跳板机架构说明

```
外网用户 ----> 跳板机（有公网IP）----> 目标服务器（只有内网IP）
                  |                    |
             公网IP: 1.2.3.4    内网IP: 192.168.1.100
             内网IP: 192.168.1.1
```

### 3.1 配置要求

#### 跳板机要求：
- 有公网 IP
- 能访问目标服务器
- 开放 SSH 端口（22）
- 配置允许 SSH 转发

#### 目标服务器要求：
- 允许跳板机 SSH 连接
- 与跳板机网络连通

#### 网络配置：
- 允许 SSH 流量
- 配置正确的路由

### 3.2 连通性测试

```bash
# 在跳板机上测试
ping 目标服务器内网IP
telnet 目标服务器内网IP 22

# 在外网机器上测试
ssh -v -J 用户名@跳板机公网IP 用户名@目标服务器内网IP
```

## 4. CI/CD 流程说明

### 4.1 CI 流程 (ci.yml)
- 运行单元测试并上传覆盖率报告
- 运行代码静态检查
- 构建 Docker 镜像并推送到 Docker Hub

### 4.2 CD 流程 (cd.yml)
- 在 CI 流程成功后自动触发
- 通过 SSH 连接到生产服务器
- 部署最新版本应用

### 4.3 部署脚本 (start.sh)
- 安装配置 Docker
- 安装配置 Prometheus
- 安装配置 Grafana
- 配置防火墙规则
- 部署 ggcache 容器

### 4.4 自动化流程
1. 推送代码到 main 分支
2. 触发 CI 流程
3. CI 成功后触发 CD 流程
4. 自动部署到生产服务器
