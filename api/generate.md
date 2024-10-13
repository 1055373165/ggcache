## 0. environment install

- go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
- go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest


## 1. preparing proto files

- Define Request and Response message as structures for rpc requests and responses
- Define the ggcache service and rpc method, taking the request structure as a parameter and the response structure as the return value


## 2. generate stub and skeleton code

Generate the specified xxx.pb.go and xxx _ grpc.pb.go files using the protoc tool.(in api direcotory)

- 先创建好 ggcachepb 和 studentpb 目录，然后在 api 目录下分别执行：

```
protoc --go_out=ggcachepb --go_opt=paths=source_relative \
--go-grpc_out=ggcachepb --go-grpc_opt=paths=source_relative \
ggcache.proto

protoc --go_out=studentpb --go_opt=paths=source_relative \
--go-grpc_out=studentpb --go-grpc_opt=paths=source_relative \
student.proto
```

其中 

--go_out 指定 ggcache.pb.go 的输出路径

--go-grpc_out 指定了 ggcache_grpc.pb.go 的输出路径


## 3. Implement rpc client and server

- client.go
- server.go