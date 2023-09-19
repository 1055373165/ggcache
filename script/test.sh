#! /bin/bash

echo ">>> start test"

cd ../grpc/rpcCallClient/

# test1
go run client.go

# test2
cd ../serviceRegisterCall/
go run client.go


