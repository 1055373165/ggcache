.PHONY: start stop test test-grpc test-http benchmark benchmark-grpc benchmark-http clean help

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

# 启动所有服务
start:
	@echo "Starting all services..."
	./start.sh

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