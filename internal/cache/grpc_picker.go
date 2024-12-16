package cache

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/1055373165/ggcache/internal/middleware/etcd/discovery/discovery3"
	"github.com/1055373165/ggcache/internal/service/consistenthash"
	"github.com/1055373165/ggcache/pkg/pb"
	"github.com/1055373165/ggcache/utils/logger"
	"google.golang.org/grpc"
)

// Server implements the GroupCache gRPC service.
type Server struct {
	pb.UnimplementedGroupCacheServer
	Addr        string                  // Server's address (host:port)
	Status      bool                   // true: running false: stop
	stopsSignal chan error             // Notify registry to stop keepalive lease
	mu          sync.Mutex             // Guards peers and clients
	peers       *consistenthash.Map    // Consistent hash map for peer selection
	clients     map[string]*Client     // Cache of gRPC clients, keyed by peer address
	update      chan bool              // Update signal for peers
}

// NewServer creates a new GroupCache gRPC server.
func NewServer(update chan bool, addr string) (*Server, error) {
	if addr == "" {
		addr = "127.0.0.1:9999"
	}

	if !validate.ValidPeerAddr(addr) {
		return nil, fmt.Errorf("invalid addr %s, expect address format is x.x.x.x:port", addr)
	}

	return &Server{
		Addr:   addr,
		update: update,
	}, nil
}

// Get handles gRPC requests to fetch values from the cache.
//
// Parameters:
//   - ctx: The request context
//   - req: The GetRequest containing group name and key
//
// Returns:
//   - *pb.GetResponse: The response containing the fetched value
//   - error: Any error encountered during the fetch
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	group, key := req.GetGroup(), req.GetKey()
	resp := &pb.GetResponse{}
	logger.LogrusObj.Infof("[GroupCache server %s] Recv RPC Request - (%s)/(%s)", s.Addr, group, key)

	if key == "" || group == "" {
		return resp, fmt.Errorf("key and group name are required")
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
	s.mu.Lock()

	if len(peersAddrs) == 0 {
		peersAddrs = []string{s.Addr}
	}

	s.peers = consistenthash.NewConsistentHash(defaultReplicas, nil)
	s.peers.AddTruthNode(peersAddrs)
	s.clients = make(map[string]*Client)

	for _, peersAddr := range peersAddrs {
		if !validate.ValidPeerAddr(peersAddr) {
			s.mu.Unlock()
			panic(fmt.Sprintf("[peer %s] invalid address format, expect x.x.x.x:port", peersAddr))
		}
		s.clients[peersAddr] = NewClient("GroupCache")
	}
	s.mu.Unlock()

	go func() {
		for {
			select {
			case <-s.update: // 节点数量发生变化（新增或者删除）整个负载均衡视图需要动态修改
				s.reconstruct()
			case <-s.stopsSignal:
				s.Stop()
			default:
				time.Sleep(time.Second * 2)
			}
		}
	}()
}

func (s *Server) reconstruct() {
	serviceList, err := discovery3.ListServicePeers("GroupCache")
	if err != nil { // 如果没有拿到服务实例列表，暂时先维持当前视图
		return
	}

	s.mu.Lock()
	s.peers = consistenthash.NewConsistentHash(defaultReplicas, nil)
	s.peers.AddTruthNode(serviceList)
	s.clients = make(map[string]*Client)

	for _, peerAddr := range serviceList {
		if !validate.ValidPeerAddr(peerAddr) {
			panic(fmt.Sprintf("[peer %s] invalid address format, expect x.x.x.x:port", peerAddr))
		}

		s.clients[peerAddr] = NewClient("GroupCache")
	}
	s.mu.Unlock()
	logger.LogrusObj.Infof("hash ring reconstruct, contain service peer %v", serviceList)
}

// PickPeer selects a peer to handle a given key.
// It returns (nil, false) if no peer was picked or if the key should be handled locally.
func (s *Server) PickPeer(key string) (Fetcher, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peerAddr := s.peers.GetTruthNode(key)

	// Pick itself
	// peerAddr 在系统刚启动时一致性视图还未构建完成
	if peerAddr == s.Addr || peerAddr == "" {
		logger.LogrusObj.Infof("oohhh! pick myself, i am %s", s.Addr)
		return nil, false
	}

	logger.LogrusObj.Infof("[current peer %s] pick remote peer: %s", s.Addr, peerAddr)
	return s.clients[peerAddr], true
}

// Start initializes and starts the gRPC server.
// It listens on the server's address and registers the GroupCache service.
func (s *Server) Start() {
	s.mu.Lock()
	if s.Status {
		s.mu.Unlock()
		fmt.Printf("server %s is already started", s.Addr)
		return
	}

	s.Status = true
	s.stopsSignal = make(chan error)

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
		err := discovery3.Register("GroupCache", s.Addr, s.stopsSignal)
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}

		close(s.stopsSignal)

		err = lis.Close()
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}
		logger.LogrusObj.Warnf("[%s] Revoke service and close tcp socket", s.Addr)
	}()
	s.mu.Unlock()

	// Serve accepts incoming connections on the listener list, creating a new Server Transport and service Goroutine for each connection.
	// The service goroutines read gRPC requests and then call the registered handlers to reply to them.
	if err := grpcServer.Serve(lis); s.Status && err != nil {
		logger.LogrusObj.Fatalf("failed to serve %s, error: %v", s.Addr, err)
		return
	}
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Status {
		return
	}

	// notify registry to stop release keepalive
	s.stopsSignal <- nil
	// change server running status
	s.Status = false
	s.clients = nil // help GC
	s.peers = nil
}
