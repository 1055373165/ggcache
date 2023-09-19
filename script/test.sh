#! /bin/bash

echo ">>> start test"

cd ../grpc/client/

# test1
go run client.go

# test2
cd ../sendRequest/
go run client.go


