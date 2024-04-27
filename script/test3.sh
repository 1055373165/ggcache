#! /bin/bash

echo  "start test 1"

go run test/grpc1/grpc_client1.go

echo  "test 1 over"

echo  "start test 2, if want stop please interrupt by signal"

go run test/grpc2/grpc_client2.go