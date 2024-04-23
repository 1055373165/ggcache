1. preparing proto files
   1. Define Request and Response message as structures for rpc requests and responses
   2. Define the GroupCache service and rpc method, taking the request structure as a parameter and the response structure as the return value

2. Generate the specified xxx.pb.go and xxx _ grpc.pb.go files using the protoc tool.(in api direcotory)

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ggcache.proto

1. Implement rpc client and server
   1. client.go
   2. server.go