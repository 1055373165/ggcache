# 服务启动方式

## localhost
1. brew install prometheus
2. brew install grafana
3. brew services start prometheus
4. brew services start grafana
5. go get github.com/prometheus/client_golang/prometheus
6. go mod tidy
7. cd deploy
8. prometheus --config.file=./prometheus/prometheus.yml

## docker
参考 start.sh