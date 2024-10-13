package internal

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"fmt"

	pb "github.com/1055373165/ggcache/api/ggcachepb"

	"google.golang.org/grpc"

	discovery "github.com/1055373165/ggcache/discovery"

	"github.com/1055373165/ggcache/utils/logger"
	"github.com/1055373165/ggcache/utils/validate"
)

var _ Picker = (*Server)(nil)

// The Server module provides communication capabilities between different groupcache instances.
// In this way, the groupcache service deployed on other machines can forward the request to the instance that owns the cached data.
// The consistent hash algorithm is responsible for finding out which host has the value corresponding to the key.
const (
	CacheServiceName = "GroupCache"
)

// Server and Group are decoupled, so Server must implement concurrency safety by itself.
type Server struct {
	pb.UnimplementedGroupCacheServer

	Addr       string        // format: ip:port
	Status     bool          // true: running false: stop
	stop       chan error    // Service termination signal channel.
	viewUpdate chan struct{} // Notification channel when consistent hashing view changes.

	mutex       sync.Mutex // Protect the following fields.
	loadBalance *ConsistentHash
	clients     map[string]*Client
}

// NewServer creates a cache server, if addr is empty, use the default service address.
func NewServer(viewUpdate chan struct{}, addr string) (*Server, error) {
	if addr == "" {
		addr = DefaultAddr
	}

	// Nodes in a distributed system that can provide the same service are considered equal and are generally called Peer.
	if !validate.ValidPeerAddr(addr) {
		return nil, fmt.Errorf("expect address format is x.x.x.x:port, but got %s", addr)
	}

	return &Server{Addr: addr, viewUpdate: viewUpdate}, nil
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

// SetPeers configures each remote host IP to the Server
func (s *Server) SetPeers(peersAddrs []string) {
	s.mutex.Lock()

	if len(peersAddrs) == 0 {
		peersAddrs = []string{s.Addr}
	}

	s.loadBalance = NewConsistentHash(DefaultReplicas, nil)
	s.loadBalance.AddTruthNode(peersAddrs)
	s.clients = make(map[string]*Client)

	for _, peersAddr := range peersAddrs {
		if !validate.ValidPeerAddr(peersAddr) {
			s.mutex.Unlock()
			panic(fmt.Sprintf("[peer %s] invalid address format, it should be x.x.x.x:port", peersAddr))
		}
		// GroupCache/localhost:9999
		// GroupCache/localhost:10000
		// GroupCache/localhost:10001
		// attention：服务发现原理建议看下 Endpoint 源码, key 是 service/addr value 是 addr
		// 服务解析时按照 service 进行前缀查询，找到所有服务节点
		// 而 clusters 前缀是为了拿到所有实例地址做一致性哈希使用的
		// 注意 service 要和你在 protocol 文件中定义的服务名称一致
		s.clients[peersAddr] = NewClient(CacheServiceName)
	}
	s.mutex.Unlock()

	go func() {
		for {
			select {
			case <-s.viewUpdate: // 节点数量发生变化（新增或者删除）整个负载均衡视图需要动态修改
				s.reconstruct()
			case <-s.stop:
				s.Stop()
			default:
				time.Sleep(time.Second * 2)
			}
		}
	}()
}

func (s *Server) reconstruct() {
	serviceList, err := discovery.ListServicePeers(CacheServiceName)
	if err != nil { // 如果没有拿到服务实例列表，暂时先维持当前视图
		return
	}

	s.mutex.Lock()
	s.loadBalance = NewConsistentHash(DefaultReplicas, nil)
	s.loadBalance.AddTruthNode(serviceList)
	s.clients = make(map[string]*Client)

	for _, peerAddr := range serviceList {
		if !validate.ValidPeerAddr(peerAddr) {
			panic(fmt.Sprintf("[peer %s] invalid address format, expect x.x.x.x:port", peerAddr))
		}

		// demo: GroupCache/127.0.0.1:9999
		s.clients[peerAddr] = NewClient(CacheServiceName)
	}
	s.mutex.Unlock()
	logger.LogrusObj.Infof("hash ring reconstruct, contain service peer %v", serviceList)
}

// Selects which cache a key request should be sent to based on a consistent hash value
func (s *Server) Pick(key string) (Fetcher, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// The real node of the hit
	peerAddr := s.loadBalance.GetTruthNode(key)

	// Pick itself
	// peerAddr 在系统刚启动时一致性视图还未构建完成
	if peerAddr == s.Addr || peerAddr == "" {
		logger.LogrusObj.Infof("oohhh! pick myself, i am %s", s.Addr)
		return nil, false
	}

	logger.LogrusObj.Infof("[current peer %s] pick remote peer: %s", s.Addr, peerAddr)
	return s.clients[peerAddr], true
}

/*
------------start service----------------
1. set server running status
2. initialize stop channel to notify registry stop keepalive lease
3. initialize tcp socket and start listening
4. register the customize rpc service to grpc, so that grpc can distribute the request to the server for processing
5. use etcd service registry which can directly obtain the client connection to the given service through the service name, that is, the grpc channel, and then create a client Stub, which implements the same method as the server and calls it directly
*/
func (s *Server) Start() {
	s.mutex.Lock()
	if s.Status {
		s.mutex.Unlock()
		fmt.Printf("server %s is already started", s.Addr)
		return
	}

	s.Status = true
	s.stop = make(chan error)

	// waiting for client to connect
	port := strings.Split(s.Addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("failed to listen %s, error: %v", s.Addr, err)
		return
	}

	// getting a customized implementation of a grpc Server instance
	grpcServer := grpc.NewServer()
	// registers the service and its implementation to a concrete type that implements GroupCacheServer interface.
	pb.RegisterGroupCacheServer(grpcServer, s)
	defer s.Stop()

	// service registry
	go func() {
		// register never return unless stop signal received (blocked)
		// for-select combination implement permanent blocking
		err := discovery.Register(CacheServiceName, s.Addr, s.stop)
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}

		close(s.stop)

		err = lis.Close()
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}
		logger.LogrusObj.Warnf("[%s] Revoke service and close tcp socket", s.Addr)
	}()
	s.mutex.Unlock()

	// Serve accepts incoming connections on the listener list, creating a new Server Transport and service Goroutine for each connection.
	// The service goroutines read gRPC requests and then call the registered handlers to reply to them.
	if err := grpcServer.Serve(lis); s.Status && err != nil {
		logger.LogrusObj.Fatalf("failed to serve %s, error: %v", s.Addr, err)
		return
	}
}

func (s *Server) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.Status {
		return
	}

	// notify registry to stop release keepalive
	s.stop <- nil
	// change server running status
	s.Status = false
	s.clients = nil // help GC
	s.loadBalance = nil
}
