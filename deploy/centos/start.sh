#!/bin/bash

# 检查是否提供了 Docker Hub 用户名
if [ -z "$1" ]; then
    echo "Please provide Docker Hub username as first argument"
    exit 1
fi

DOCKERHUB_USERNAME=$1

# 安装必要的软件包
yum update -y
yum install -y yum-utils device-mapper-persistent-data lvm2
yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
yum install -y docker-ce docker-ce-cli containerd.io

# 启动 Docker
systemctl start docker
systemctl enable docker

# 安装 Prometheus
useradd --no-create-home --shell /bin/false prometheus
mkdir -p /etc/prometheus
mkdir -p /var/lib/prometheus

# 下载最新版本的 Prometheus
PROMETHEUS_VERSION=$(curl -s https://api.github.com/repos/prometheus/prometheus/releases/latest | grep tag_name | cut -d '"' -f 4)
wget https://github.com/prometheus/prometheus/releases/download/${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION#v}.linux-amd64.tar.gz
tar xvf prometheus-*.tar.gz
cp prometheus-*/prometheus /usr/local/bin/
cp prometheus-*/promtool /usr/local/bin/
chown prometheus:prometheus /usr/local/bin/prometheus
chown prometheus:prometheus /usr/local/bin/promtool
chown -R prometheus:prometheus /etc/prometheus
chown -R prometheus:prometheus /var/lib/prometheus

# 创建 Prometheus systemd 服务
cat > /etc/systemd/system/prometheus.service << EOF
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \
    --config.file /etc/prometheus/prometheus.yml \
    --storage.tsdb.path /var/lib/prometheus/ \
    --web.console.templates=/etc/prometheus/consoles \
    --web.console.libraries=/etc/prometheus/console_libraries

[Install]
WantedBy=multi-user.target
EOF

# 安装 Grafana
cat > /etc/yum.repos.d/grafana.repo << EOF
[grafana]
name=grafana
baseurl=https://packages.grafana.com/oss/rpm
repo_gpgcheck=1
enabled=1
gpgcheck=1
gpgkey=https://packages.grafana.com/gpg.key
sslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt
EOF

yum install -y grafana

# 启动服务
systemctl daemon-reload
systemctl start prometheus
systemctl enable prometheus
systemctl start grafana-server
systemctl enable grafana-server

# 配置防火墙
firewall-cmd --permanent --add-port=9090/tcp  # Prometheus
firewall-cmd --permanent --add-port=3000/tcp  # Grafana
firewall-cmd --permanent --add-port=9999/tcp  # ggcache
firewall-cmd --permanent --add-port=10000/tcp # ggcache
firewall-cmd --permanent --add-port=10001/tcp # ggcache
firewall-cmd --reload

# 拉取并运行 ggcache 容器
docker pull ${DOCKERHUB_USERNAME}/ggcache:latest
docker run -d \
    --name ggcache \
    -p 9999:9999 \
    -p 10000:10000 \
    -p 10001:10001 \
    -p 2222:2222 \
    -p 2223:2223 \
    -p 2224:2224 \
    -p 6060:6060 \
    -p 6061:6061 \
    -p 6062:6062 \
    ${DOCKERHUB_USERNAME}/ggcache:latest