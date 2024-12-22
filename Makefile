.PHONY: start stop test test-grpc test-http benchmark benchmark-grpc benchmark-http clean help docker-build docker-start docker-stop docker-logs docker-shell

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
help:
	@echo "Available commands:"
	@echo "  make start          - Start all services (etcd, docker, servers)"
	@echo "  make stop           - Stop all running services"
	@echo "  make test-grpc      - Run basic gRPC client test"
	@echo "  make test-http      - Run basic HTTP client test"
	@echo "  make benchmark-grpc - Run gRPC benchmark test"
	@echo "  make benchmark-http - Run HTTP benchmark test"
	@echo "  make clean          - Clean up temporary files and logs"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-start   - Start all services in Docker"
	@echo "  make docker-stop    - Stop all Docker services"
	@echo "  make docker-logs    - Show Docker logs"
	@echo "  make docker-shell   - Enter ggcache container"

# 本地启动
start:
	@echo "Starting all services..."
	./start.sh

# Docker 相关命令
# 构建 Docker 镜像
docker-build:
	@echo "Building Docker image..."
	docker build -t ggcache .

# 启动所有 Docker 服务
docker-start:
	@echo "Starting all services in Docker..."
	docker-compose up -d
	@echo "Waiting for services to start..."
	sleep 5
	@echo "Showing ggcache logs and entering container..."
	docker-compose logs ggcache

# 进入 ggcache 容器
docker-shell:
	@echo "Entering ggcache container..."
	docker-compose exec ggcache bash

# 停止所有 Docker 服务
docker-stop:
	@echo "Stopping all Docker services..."
	docker-compose down

# 查看服务日志
docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f

# 停止所有服务
stop:
	@echo "Stopping all services..."
	./stop.sh

# 运行基本 gRPC 测试
test-grpc:
	@echo "Running gRPC test..."
	./test/grpc/run_clients.sh

# 运行基本 HTTP 测试
test-http:
	@echo "Running HTTP test..."
	./test/http/http_test1.sh

# 清理
clean:
	@echo "Cleaning up..."
	docker system prune -f
	rm -f *.log
	rm -rf test/benchmark/results_*
	rm -rf test/grpc/logs
	@echo "Cleanup complete"