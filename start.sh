#!/bin/bash

# Function to check if a process is running on a specific port
check_port() {
    lsof -i :$1 > /dev/null 2>&1
    return $?
}

# Function to wait for a port to be available
wait_for_port() {
    local port=$1
    local retries=30
    while [ $retries -gt 0 ]; do
        if check_port $port; then
            return 0
        fi
        retries=$((retries-1))
        sleep 1
    done
    return 1
}

echo "Starting etcd cluster..."
goreman -f pkg/etcd/cluster/Procfile start &

echo "Starting docker containers..."
docker-compose up -d

# Wait a bit for services to initialize
sleep 5

echo "Starting server on port 9999..."
go run main.go -port 9999 &

# Wait for the first server to start
sleep 3

echo "Starting server on port 10000..."
go run main.go -port 10000 -metricsPort 2223 -pprofPort 6061 &

# Wait for the second server to start
sleep 3

echo "Starting server on port 10001..."
go run main.go -port 10001 -metricsPort 2224 -pprofPort 6062 &

# Wait for the third server to start
sleep 3

echo "Starting clients..."
./test/grpc/run_clients.sh

# Keep the script running
wait
