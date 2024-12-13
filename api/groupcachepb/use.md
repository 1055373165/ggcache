1. 编写 proto 文件
   1. 定义 Request 和 Response message 作为 rp 从请求和响应的结构体
   2. 定义 GroupCache 服务，Get 方法，以请求结构体作为参数，以响应结构体作为返回值

2. 使用 protoc 工具生成指定的 xxx.pb.go 和 xxx_grpc.pb.go 文件

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative groupcachepb/groupcache.proto

3. rpc 客户端和服务端启动文件
   1. client.go
   2. server.go