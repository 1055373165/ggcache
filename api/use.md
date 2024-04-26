1. preparing proto files
   1. Define Request and Response message as structures for rpc requests and responses
   2. Define the ggcache service and rpc method, taking the request structure as a parameter and the response structure as the return value

2. Generate the specified xxx.pb.go and xxx _ grpc.pb.go files using the protoc tool.(in api direcotory)

- 先创建好 ggcache 和 student 目录

- protoc --go_out=./ggcache --go_opt=paths=source_relative --go-grpc_out=./ggcache --go-grpc_opt=paths=source_relative ggcache.proto

- protoc --go_out=./student --go_opt=paths=source_relative --go-grpc_out=./student --go-grpc_opt=paths=source_relative student.proto

其中 --go_out 指定 ggcache.pb.go 的输出路径，--go-grpc_out 指定了 ggcache_grpc.pb.go 的输出路径


3. Implement rpc client and server
   1. client.go
   2. server.go