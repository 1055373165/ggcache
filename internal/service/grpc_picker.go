package service

import (
	"context"

	"fmt"
	"net"
	"strings"
	"sync"

	"google.golang.org/grpc"

	pb "github.com/1055373165/ggcache/api/groupcachepb"

	"github.com/1055373165/ggcache/internal/service/consistenthash"
	"github.com/1055373165/ggcache/utils/logger"
	"github.com/1055373165/ggcache/utils/validate"
)

// 测试 Server 是否实现了 Picker 接口
var _ Picker = (*Server)(nil)

// The Server module provides communication capabilities between groupcache.
// In this way, groupcache deployed on other machines can obtain the cache by accessing the server.
// As for which host to find, consistent hashing is responsible.
const (
	defaultAddr     = "127.0.0.1:9999"
	defaultReplicas = 50
)

// Server and Group are decoupled, so the server must implement concurrency control by itself
type Server struct {
	pb.UnimplementedGroupCacheServer

	Addr        string     // format: ip:port
	Status      bool       // true: running false: stop
	stopsSignal chan error // 通知 registery revoke 服务
	mu          sync.Mutex
	consHash    *consistenthash.ConsistentHash
	clients     map[string]*Client
}

// New Server creates a cache server. If addr is empty, default Addr is used.
func NewServer(addr string) (*Server, error) {
	if addr == "" {
		addr = defaultAddr
	}

	if !validate.ValidPeerAddr(addr) {
		return nil, fmt.Errorf("invalid addr %s, it should be x.x.x.x:port", addr)
	}
	return &Server{Addr: addr}, nil
}

// Server Implement GroupCacheServer in groupcachepb
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	group, key := req.GetGroup(), req.GetKey()
	resp := &pb.GetResponse{}
	logger.LogrusObj.Infof("[Groupcache server %s] Recv RPC Request - (%s)/(%s)", s.Addr, group, key)

	if key == "" || group == "" {
		return resp, fmt.Errorf("key and group name is reqiured")
	}

	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("group %s not found", group)
	}

	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}

	resp.Value = view.Bytes()
	return resp, nil
}

/*
------------启动服务----------------
1. 设置 status = true 表示服务器已经在运行
2. 初始化 stop channel，用于通知 registry stop keepalive
3. 初始化 tcp socket 并开始监听指定端口
4. 注册自定义 grpc 服务至 grpc 空白实例，这样 grpc 收到 request 可以分发给 server 处理
5. 服务注册（阻塞），异步完成
*/
func (s *Server) Start() error {
	s.mu.Lock()

	if s.Status {
		s.mu.Unlock()
		return fmt.Errorf("server %s is already started", s.Addr)
	}

	s.Status = true
	s.stopsSignal = make(chan error)

	port := strings.Split(s.Addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen %s, error: %v", s.Addr, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGroupCacheServer(grpcServer, s)

	go func() {
		// This operation is blocking and therefore completed asynchronously in a goroutine
		err := Register("GroupCache", s.Addr, s.stopsSignal, s)
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}
		// When Register exits, the service stops
		close(s.stopsSignal)
		err = lis.Close()
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}
	}()

	s.mu.Unlock()
	if err := grpcServer.Serve(lis); s.Status && err != nil {
		return fmt.Errorf("failed to serve %s, error: %v", s.Addr, err)
	}
	return nil
}

// SetPeers configures each remote host IP to the Server
func (s *Server) SetPeers(peersAddr []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.consHash = consistenthash.NewConsistentHash(defaultReplicas, nil)
	s.consHash.AddTruthNode(peersAddr)
	s.clients = make(map[string]*Client)

	for _, peersAddr := range peersAddr {
		if !validate.ValidPeerAddr(peersAddr) {
			panic(fmt.Sprintf("[peer %s] invalid address format, it should be x.x.x.x:port", peersAddr))
		}

		// GroupCache/localhost:9999
		// GroupCache/localhost:10000
		// GroupCache/localhost:10001
		// attention：服务发现原理建议看下 Endpoint 源码, key 是 service/addr value 是 addr
		// 服务解析时按照 service 进行前缀查询，找到所有服务节点
		// 而 clusters 前缀是为了拿到所有实例地址做一致性哈希使用的
		// 注意 service 要和你在 protocol 文件中定义的服务名称一致
		s.clients[peersAddr] = NewClient("GroupCache")
	}
}

// Selects which cache a key request should be sent to based on a consistent hash value
func (s *Server) Pick(key string) (Fetcher, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// The real node of the hit
	peerAddr := s.consHash.GetTruthNode(key)

	// Pick itself
	if peerAddr == s.Addr {
		logger.LogrusObj.Infof("oohhh! pick myself, i am %s\n", s.Addr)
		return nil, false
	}

	logger.LogrusObj.Infof("[current peer %s] pick remote peer: %s\n", s.Addr, peerAddr)
	return s.clients[peerAddr], true
}

// Stop stops the server running
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Status {
		return
	}
	// Sends a signal to stop keepAlive hearbeat
	s.stopsSignal <- nil
	s.Status = false
	// Clear consistent hash information to help GC perform garbage collection
	s.clients = nil
	s.consHash = nil
}
