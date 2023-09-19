# 实现基于 protobuf 协议的 RPC 通信

geecache 并没有实现真正的 RPC 通信，只是简单对请求和响应使用 protobuf 序列化了一下而已，因此本部分尝试实现基于 protobuf 的 RPC 通信。

这只是第一版分析，后面基于 etcd 实现时，对 gRPC 流程进一步进行了详细的分析；

基本流程如下：

## 1. 编写 proto 文件


```protobuf
syntax="proto3";
package xxxx;
option go_package = "./groupcachepb" // 指定包导入路径，必须
```

- Request

```protobuf
message Request {
  string groupname = 1;
  string key = 2;
}
```

- Response

```protobuf
message Response {
  bytes value = 1;
}
```

- 定义服务

```protobuf
service GroupCache {
    rpc Get(Request) returns (Response) {}
}
```

## 安装 protobuf 编译器

1. go install google.golang.org/protobuf/cmd/[protoc-gen-go@v1.28](https://mailto:protoc-gen-go@v1.28) 
2. go install google.golang.org/grpc/cmd/[protoc-gen-go-grpc@v1.2](https://mailto:protoc-gen-go-grpc@v1.2) 

## 生成 protobuf 文件

- xxx.pb.go
- xxx_grpc.pb.go

```bash
protoc --go_out=./groupcachepb --go_opt=paths=source_relative \
--go-grpc_out=./groupcachepb --go-grpc_opt=paths=source_relative \
groupcache.proto
```

- 使用 --go_out 指定生成 ***.pb.go 的文件目录
- 使用  --go_opt=paths=source_relative 指定为相对路径（相对于当前所在的目录）
- 使用 --go-grpc_out 指定生成 ***_grpc.pb.go 文件的目录
- 使用 --go-grpc_opt=paths=source_relative 相对于当前所在路径下
- ***.proto 指定为当前目录下的哪一个 proto 文件生成 pb 文件

### xxx.pb.go 文件

- 主要包括了两个结构体以及序列化和反序列化方法

```go
  // 定义请求消息体
type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	**Groupname string `protobuf:"bytes,1,opt,name=groupname,proto3" json:"groupname,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`**
}
```

```go
// 定义响应消息体
type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	**Value []byte `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`**
}
```

### xxx_grpc.pb.go 文件

主要包含了 grpc 的客户端和服务端的定义

- GroupCacheClient 是 GroupCache 服务的客户端 API。（这个是根据我们定义生成的）

```go
type GroupCacheClient interface {
	Get(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}
```

- grpc 的客户端连接接口
ClientConnInterface 定义了客户端执行单向和流 RPC 所需的功能。它由 ClientConn 实现，只供生成的代码引用。

提供了单向 RPC 和流式 RPC

```go
type groupCacheClient struct {
	cc grpc.ClientConnInterface
}

type ClientConnInterface interface {
    // Invoke执行一元RPC，并在收到应答后返回。
    Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...CallOption) error
    // NewStream 开始执行流 RPC。
    NewStream(ctx context.Context, desc *StreamDesc, method string, opts ...CallOption) (ClientStream, error)
}
```

看下 CallOption：callloption 在调用启动前配置 Call，或在调用完成后从调用中提取信息。

```go
type CallOption interface {
	// before 会在调用发送到任何服务器之前被调用。
	before(*callInfo) error

	// 在调用完成后调用。After无法返回错误，因此任何失败都应通过输出参数报告
	after(*callInfo, *csAttempt)
}
```

callInfo 包含 RPC 的所有相关配置和信息。

```go
type callInfo struct {
	compressorType        string
	failFast              bool
	maxReceiveMessageSize *int
	maxSendMessageSize    *int
	creds                 credentials.PerRPCCredentials
	contentSubtype        string
	codec                 baseCodec
	maxRetryRPCBufferSize int
	onFinish              []func(err error)
}
```

- compressorType：压缩类型
- 快速失败
- 最大接收数据大小
- 最大发送数据大小
- PerRPCCredentials 为需要将安全信息附加到每个 RPC 的凭证（如 oauth2）定义了通用接口
- 数据内容的子类型
- baseCodec包含编解码器和编码的功能。但省略了名称/字符串，它们在两者之间有所不同，除了编码包中的注册表之外，其他任何内容都不需要；

```go
type baseCodec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}
```

- maxRetryRPCBufferSize 最大重试 RPC 缓冲区大小
- 调用过程中产生的所有错误切片


- 新建服务客户端

```go
func NewGroupCacheClient(cc grpc.ClientConnInterface) GroupCacheClient {
	return &groupCacheClient{cc}
}
```

因为 groupCacheClient 实现了接口 GroupCacheClient 因此可以赋值给接口 GroupCacheClient；

那么这个返回的对象就拥有了 groupCacheClient 的 Get 能力

```go
func (c *groupCacheClient) Get(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
    // 实际发出调用的地方
	err := c.cc.Invoke(ctx, "/groupcachepb.GroupCache/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
} 
```

1. 新建用于返回给客户端的输出结构体
2. RPC 调用，参数为
	1. 上下文对象
	2. 调用路径
	3. 请求
	4. 响应
	5. 调用操作
		1. 调用之前的操作 before
		2. 调用之后的操作 after
3. 如果执行成功，err 为 nil，返回响应和 nil
4. 否则，返回 nil 和错误给调用方。

核心函数 grpc.ClientConnInterface.invoke()，是接口的一个方法；

```go
Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...CallOption) error
```

ClientConn 实现了该接口，我们看下 ClientConn，它才是实际负责调用的类型；

ClientConn 表示与概念端点的**虚拟连接**，用于执行 RPC。一个 ClientConn 可以根据配置、负载等情况，自由地与端点建立零个或多个实际连接。

它还可以自由决定使用哪个实际端点，并可在每次 RPC 时更改，从而实现客户端负载平衡。

ClientConn 封装了一系列功能，包括名称解析、TCP 连接建立（重试和后退）和 TLS 握手。它还能通过重新解析名称和重新连接来处理已建立连接上的错误。

```go
type ClientConn struct {
        // 在 Dial 时使用 Backgroud() 上下文进行初始化。
	ctx    context.Context    
        // 关闭时取消。
	cancel context.CancelFunc 

	// 以下内容在 Dial 时初始化，之后只读
  	// User's dial target. 用户的拨号目标。
	target          string               
	parsedTarget    resolver.Target      // See parseTargetAndFindResolver().
	authority       string               // See determineAuthority().
	// 默认和用户指定的拨号选项。
	dopts           dialOptions          // Default and user specified dial options.
	// 标识符是一个不透明标识符，可在 channelz 数据库中唯一标识一个实体
	channelzID      *channelz.Identifier // Channelz identifier for the channel.
	// 生成器会创建一个解析器，用于监视名称解析更新。(注册解析器)
	resolverBuilder resolver.Builder     // See parseTargetAndFindResolver().
	// ccBalancerWrapper位于ClientConn和Balancer之间。
	balancerWrapper *ccBalancerWrapper   // Uses gracefulswitch.balancer underneath.
        // idlenessManager定义跟踪通道上RPC活动所需的功能。
	idlenessMgr     idlenessManager

	// The following provide their own synchronization, and therefore don't
	// require cc.mu to be held to access them.
	// 以下内容提供了自己的同步功能，因此无需持有 cc.mu 即可访问。
	// connectivityStateManager 保存 ClientConn 的连接状态。
	// 该结构最终将被导出，以便均衡器可以访问它。
	csMgr              *connectivityStateManager
	// pickerWrapper 是 balancer.Picker 的包装器。
	// 它会阻止某些选取操作，并在有选取器更新时解除阻止。
	blockingpicker     *pickerWrapper
	// 安全配置选择器允许安全切换配置选择器实现，以便保证更新配置选择器返回时以前的值不会被使用。
	safeConfigSelector iresolver.SafeConfigSelector
	// channelzData 用于为 ClientConn、addrConn 和 Server 存储 channelz 相关数据。
	// 这些字段不能嵌入原始结构体（如 ClientConn）中，因为要在 32 位机器上对 int64 变量进行原子操作，
	// 用户有责任执行内存对齐。在这里，通过将这些 int64 字段分组到一个结构体中，我们可以强制对齐。
	czData             *channelzData
	// Value提供了一致类型值的原子加载和存储。
	// value的零值从Load返回nil。一旦Store被调用，Value就不能被复制。
	retryThrottler     atomic.Value // Updated from service config.从服务配置更新。

	// firstResolveEvent is used to track whether the name resolver sent us at
	// least one update. RPCs block on this event.
	// Event表示将来可能发生的一次性事件。
	// firstResolveEvent 用于跟踪名称解析器是否至少向我们发送了一次更新。RPC 在此事件上阻塞。
	firstResolveEvent *grpcsync.Event

	// mu protects the following fields.
	// TODO: split mu so the same mutex isn't used for everything.
	mu              sync.RWMutex
	// ccResolverWrapper 是解析器的 cc 封装。它实现了 resolver.ClientConn 接口。
	resolverWrapper *ccResolverWrapper         // 在Dial中初始化；关闭时清除。
	sc              *ServiceConfig             // Latest service config received from the resolver. 从解析器收到的最新服务配置。
	// 关闭时设置为nil。
	// addrConn 是指向给定地址的网络连接。
	conns           map[*addrConn]struct{}     // Set to nil on close.
	mkp             keepalive.ClientParameters // May be updated upon receipt of a GoAway.
	idlenessState   ccIdlenessState            // Tracks idleness state of the channel.
	exitIdleCond    *sync.Cond                 // Signalled when channel exits idle.

	lceMu               sync.Mutex // protects lastConnectionError
	lastConnectionError error
}
```

- Target 结构体

Target表示gRPC的目标，如:https://github.com/grpc/grpc/blob/master/doc/naming.md中指定的。

它是从用户传递给Dial或DialContext的目标字符串中解析出来的。gRPC将其传递给解析器和平衡器。

如果目标遵循命名规范，并且解析的方案已注册到 gRPC，我们将根据规范解析目标字符串。如果目标不包含方案或者解析的方案未注册（即没有相应的解析器可用于解析端点），我们将应用默认方案，并尝试重新解析它。

Examples:

```
//   - "dns://some_authority/foo.bar"
//     Target{Scheme: "dns", Authority: "some_authority", Endpoint: "foo.bar"}
//   - "foo.bar"
//     Target{Scheme: resolver.GetDefaultScheme(), Endpoint: "foo.bar"}
//   - "unknown_scheme://authority/endpoint"
//     Target{Scheme: resolver.GetDefaultScheme(), Endpoint: "unknown_scheme://authority/endpoint"}
```

ccBalancerWrapper

ccBalancerWrapper 实现了与 balancer.Balancer 接口上的方法相对应的方法。ClientConn 可自由并发调用这些方法，而 ccBalancerWrapper 可确保从 ClientConn 到 Balancer 的**调用按顺序同步**进行。它调用 ClientConn 上未导出的方法来处理这些来自平衡器的调用。

keepalive.ClientParameters

客户端参数用于设置客户端的保活参数。这些配置客户端如何**主动探测以注意到连接何时断开并发送 ping，以便中介了解连接的活跃度**。确保这些参数的设置与服务器上的保活策略相协调，因为不兼容的设置可能会导致连接关闭。

sync.Cond

Cond 实现了一个条件变量，它是等待或宣布事件发生的 goroutines 的集合点。每个 Cond 都有一个相关的锁定器 L（通常是一个 Mutex 或 RWMutex），在改变条件和调用 Wait 方法时必须保持该锁定器。Cond 首次使用后不得复制。用 Go 内存模型的术语来说，Cond 可以使对 Broadcast 或 Signal 的调用 "先于 "其解锁的任何 Wait 调用而 "同步"。

```go
func (cc *ClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...CallOption) error {
	if err := cc.idlenessMgr.onCallBegin(); err != nil {
		return err
	}
	defer cc.idlenessMgr.onCallEnd()

	// allow interceptor to see all applicable call options, which means those
	// configured as defaults from dial option as well as per-call options
	opts = combine(cc.dopts.callOptions, opts)

	if cc.dopts.unaryInt != nil {
		return cc.dopts.unaryInt(ctx, method, args, reply, cc, invoke, opts...)
	}
	return invoke(ctx, method, args, reply, cc, opts...)
}
```

1. 在实际发起 rpc 调用前，首先调用 idlenessManager 的 onCallBegin 方法，设置跟踪通道上 RPC 活动所需的功能。
2. 使用 defer 调用 idlenessManager 的 onCallEnd 方法，在调用结束时关闭对通道上 RPC 活动的跟踪功能；
3. 然后开启允许拦截器查看所有适用的拨号选项，即那些配置为默认的拨号选项以及自定义的其他呼叫选项（dialOptions）
4. 判断是否开启了一元 RPC 拦截器，如果开启了委托给它进行 RPC 处理。
5. 否则，调用流式 invoker 进行处理。


cc.dialoptions.unaryInt != nil 相当于配置了一元 RPC 拦截器，它将接收所有一元 RPC 的调用委托，并通过调用 invoker 来完成 RPC 的处理。

UnaryClientInterceptor 可拦截客户端执行的一元 RPC。在创建 ClientConn 时，可以使用 WithUnaryInterceptor() 或 WithChainUnaryInterceptor() 将一元拦截器指定为 DialOption。当在 ClientConn 上设置了一元拦截器后，gRPC 会将所有一元 RPC 调用委托给拦截器，拦截器有责任调用 invoker 来完成 RPC 的处理。

combine

我们没有使用append，因为o1可能有额外的容量，其元素将被覆盖，这可能导致并发调用之间无意的共享(和竞争条件)

```go
func combine(o1 []CallOption, o2 []CallOption) []CallOption {
	if len(o1) == 0 {
		return o2
	} else if len(o2) == 0 {
		return o1
	}
	ret := make([]CallOption, len(o1)+len(o2))
	copy(ret, o1)
	copy(ret[len(o1):], o2)
	return ret
}
```

# 核心流式 invoke

SendMsg 通常由生成的代码调用。如果出错，SendMsg 会终止数据流。如果错误是由客户端生成的，则直接返回状态；否则返回 io.EOF，并可使用 RecvMsg 发现流的状态。

SendMsg 会阻塞，直到

- 有足够的流量控制来调度 m （请求）与传输，（直到足够的流量取走请求处理）

- 数据流已完成

- 数据流中断

SendMsg 不会等到服务器收到消息。提前的流关闭可能会导致消息丢失。

为了确保交付，用户应使用 RecvMsg 确保 RPC 成功完成。

即 SendMsg 和 RecvMsg 应该是并发运行的。

在同一个流上同时有一个 goroutine 调用 SendMsg 和另一个 goroutine 调用 Recv Msg 是安全的，但在不同的 goroutine 中对同一流调用 Send Msg 是不安全的。同时调用 Close Send 和 Send Msg 也是不安全的。

在同一个流上，SendMsg 和 RecvMsg 是并发安全的。

在调用SendMsg之后修改消息是不安全的。跟踪库和统计处理程序可能会惰性地使用消息。


RecvMsg 会一直阻塞，直到接收到一条信息进入 m （成功接收响应）或数据流结束。当信息流成功完成时，它会返回 io.EOF。如果出现其他错误，数据流将被中止，错误信息中包含 RPC 状态。

[View on canvas](https://app.eraser.io/workspace/71sCQyI8nbKOaI3CkhWj?elements=kvvUC8FCbAkjhVjzgzIqQw) 

```go
func invoke(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, opts ...CallOption) error {
	cs, err := newClientStream(ctx, unaryStreamDesc, cc, method, opts...)
	if err != nil {
		return err
	}
	if err := cs.SendMsg(req); err != nil {
		return err
	}
	return cs.RecvMsg(reply)
}
```

1. 根据参数新建一个流式 rpc 客户端
2. 调用流式客户端的 SendMsg 方法发送请求
3. 调用流式客户端的 RecvMsg 方法接收响应


# GroupCacheServer

GroupCacheServer 是 GroupCache 服务的服务器 API。为了向前兼容，所有实现都必须嵌入 UnimplementedGroupCacheServer。

```go
type GroupCacheServer interface {
	Get(context.Context, *Request) (*Response, error)
	mustEmbedUnimplementedGroupCacheServer()
}

// UnimplementedGroupCacheServer must be embedded to have forward compatible implementations.
type UnimplementedGroupCacheServer struct {
}

func (UnimplementedGroupCacheServer) Get(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedGroupCacheServer) mustEmbedUnimplementedGroupCacheServer() {}
```

可见 UnimplementedGroupCacheServer 实现了接口 GroupCacheServer

可以嵌入UnSafeGroupCacheServer以选择退出此服务的向前兼容性。不建议使用此接口，因为向GroupCacheServer添加方法会导致编译错误。

```go
type UnsafeGroupCacheServer interface {
	mustEmbedUnimplementedGroupCacheServer()
}
```

## 核心方法之 RegisterGroupCacheServer

```go
func RegisterGroupCacheServer(s grpc.ServiceRegistrar, srv GroupCacheServer) {
	s.RegisterService(&GroupCache_ServiceDesc, srv)
}
```

ServiceRegistrar 封装了一个支持服务注册的方法。它使用户能够将 grpc.Server 以外的具体类型传递给 IDL 生成代码导出的服务注册方法。

```go
type ServiceRegistrar interface {
    RegisterService(desc *ServiceDesc, impl interface{})
}
```

RegisterService将服务及其实现注册到**实现此接口的具体类型**。一旦服务器开始服务，它可能不会被调用。Desc描述服务及其方法和处理程序。Impl是传递给方法处理程序的服务实现。

比如 

```go
s.RegisterService(&GroupCache_ServiceDesc, srv)
```

其中 GroupCache_ServiceDesc 是服务的描述，而 srv(GroupCacheServer) 是服务接口的具体实现类型

ServiceDesc 表示RPC服务的规范。

```go
type ServiceDesc struct {
	ServiceName string
	// 指向服务接口的指针。用于检查用户提供的实现是否满足接口要求。
	HandlerType interface{}
	Methods     []MethodDesc
	Streams     []StreamDesc
	Metadata    interface{}
}
```

## MethodDesc

```go
type MethodDesc struct {
	MethodName string
	Handler    methodHandler
}

type methodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor UnaryServerInterceptor) (interface{}, error)
```

UnaryServerInterceptor 提供了一个拦截服务器上执行一元 RPC 的钩子。handler 是服务方法实现的包装器。拦截器有责任调用处理程序来完成 RPC。

## StreamDesc

```go
type StreamDesc struct {
	// StreamName 和 Handler 仅在服务器上注册处理程序时使用。
	StreamName string        // the name of the method excluding the service
	Handler    StreamHandler // the handler called for the method

      // ServerStreams 和 ClientStreams 用于在服务器上注册处理程序，
      // 以及在传递给 NewClientStream 和 ClientConn.NewStream 时定义 RPC 行为。
      // 至少有一个必须为 true。
	ServerStreams bool // 表示服务器可以进行流式发送
	ClientStreams bool // 表示客户端可以进行流式发送
}
```

# clientStream

ClientStream 定义了流 RPC 的客户端行为。ClientStream 方法返回的所有错误都与 status 包兼容。

```go
type ClientStream interface {
	// Header 返回从服务器接收到的标头元数据（如果有）。如果元数据尚未准备好读取，它会阻塞。
	Header() (metadata.MD, error)
	// Trailer returns the trailer metadata from the server, if there is any.
	// It must only be called after stream.CloseAndRecv has returned, or
	// stream.Recv has returned a non-nil error (including io.EOF).
	Trailer() metadata.MD
	
	// CloseSend 关闭数据流的发送方向。当遇到非零错误时，它会关闭数据流。
	// 与 SendMsg 同时调用 CloseSend 也不安全。
	CloseSend() error
	
	// Context 返回该数据流的上下文。
	// 在 Header 或 RecvMsg 返回后才可调用。
	// 一旦调用，后续的客户端重试将被禁用。
	Context() context.Context
	
	SendMsg(m interface{}) error
	RecvMsg(m interface{}) error
}
```