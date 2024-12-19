package cache

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/pkg/common/logger"
	"github.com/1055373165/ggcache/pkg/common/validate"
	"github.com/1055373165/ggcache/pkg/etcd/discovery"
	"google.golang.org/grpc"
)

var _ Picker = (*Server)(nil)

var (
	defaultAddr     = "127.0.0.1:9999"
	defaultReplicas = 50
	serviceName     = "GroupCache"
)

// Server provides gRPC-based peer-to-peer communication for distributed caching.
type Server struct {
	pb.UnimplementedGroupCacheServer

	addr        string
	isRunning   bool
	stopSignal  chan error
	updateChan  chan struct{}
	mu          sync.RWMutex
	consistHash *ConsistentMap
	clients     map[string]*Client
}

// NewServer creates a new cache server.
// If addr is empty, defaultAddr is used.
func NewServer(update chan struct{}, addr string) (*Server, error) {
	if addr == "" {
		addr = defaultAddr
	}

	if !validate.ValidPeerAddr(addr) {
		return nil, fmt.Errorf("invalid peer address %s", addr)
	}

	return &Server{addr: addr, updateChan: update}, nil
}

// Get handles gRPC requests to fetch values from the cache.
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	group, key := req.GetGroup(), req.GetKey()
	resp := &pb.GetResponse{}

	logger.LogrusObj.Infof("[Server %s] Received RPC request - group: %s, key: %s", s.addr, group, key)

	if key == "" || group == "" {
		return resp, fmt.Errorf("key and group name are required")
	}

	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("no such group: %s", group)
	}

	value, err := g.Get(key)
	if err != nil {
		return resp, err
	}

	resp.Value = value.Bytes()
	return resp, nil
}

// SetPeers configures each remote host IP to the Server
func (s *Server) SetPeers(peersAddrs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(peersAddrs) == 0 {
		peersAddrs = []string{s.addr}
	}

	s.consistHash = NewConsistentHash(defaultReplicas, nil)
	s.consistHash.AddNodes(peersAddrs...)
	s.clients = make(map[string]*Client)

	for _, peersAddr := range peersAddrs {
		if !validate.ValidPeerAddr(peersAddr) {
			s.mu.Unlock()
			panic(fmt.Sprintf("[peer %s] invalid address format, it should be x.x.x.x:port", peersAddr))
		}
		s.clients[peersAddr] = NewClient("GroupCache")
	}

	go func() {
		for {
			select {
			case <-s.updateChan:
				s.reconstruct()
			case <-s.stopSignal:
				s.Stop()
			default:
				time.Sleep(2 * time.Second)
			}
		}
	}()
}

func (s *Server) reconstruct() {
	serviceList, err := discovery.ListServicePeers("GroupCache")
	if err != nil {
		return
	}

	// 创建新的 map 和 hash 环
	newClients := make(map[string]*Client)
	newHash := NewConsistentHash(defaultReplicas, nil)
	newHash.AddNodes(serviceList...)

	// 复用现有的有效连接
	s.mu.RLock()
	for _, peerAddr := range serviceList {
		if !validate.ValidPeerAddr(peerAddr) {
			s.mu.RUnlock()
			panic(fmt.Sprintf("[peer %s] invalid address format, expect x.x.x.x:port", peerAddr))
		}

		// 尝试复用现有连接
		if client, exists := s.clients[peerAddr]; exists {
			newClients[peerAddr] = client
		} else {
			newClients[peerAddr] = NewClient("GroupCache")
		}
	}
	s.mu.RUnlock()

	// 原子替换
	s.mu.Lock()
	oldClients := s.clients
	s.clients = newClients
	s.consistHash = newHash
	s.mu.Unlock()

	for addr, client := range oldClients {
		if _, exists := newClients[addr]; !exists {
			client.Close()
		}
	}

	logger.LogrusObj.Infof("hash ring reconstruct, contain service peer %v", serviceList)
}

// Pick selects which cache node should handle the given key.
// It returns (nil, false) only when the hash ring is not yet initialized (peerAddr is empty).
// When the key is mapped to the current node, it still returns (nil, false) but this is an
// expected case indicating the key should be handled locally.
func (s *Server) Pick(key string) (Fetcher, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	peerAddr := s.consistHash.GetNode(key)
	if peerAddr == "" {
		logger.LogrusObj.Warnf("hash ring not initialized yet, handling key %s locally", key)
		return nil, false
	}

	if peerAddr == s.addr {
		logger.LogrusObj.Debugf("key %s is mapped to current node %s (local handling)", key, s.addr)
		return nil, false
	}

	logger.LogrusObj.Debugf("key %s is mapped to remote peer %s", key, peerAddr)
	return s.clients[peerAddr], true
}

// Start initializes and starts the gRPC server.
// It handles service registration, gRPC server setup, and connection management.
// Returns an error if the server fails to start or is already running.
func (s *Server) Start() error {
	if err := s.initServer(); err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	lis, err := s.setupListener()
	if err != nil {
		return fmt.Errorf("failed to setup listener: %w", err)
	}

	grpcServer := s.setupGRPCServer()

	// Start service registration in background
	errChan := make(chan error, 1)
	go s.registerService(lis, errChan)

	// Start serving requests
	if err := s.serveRequests(grpcServer, lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) initServer() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("server %s is already running", s.addr)
	}

	s.isRunning = true
	s.stopSignal = make(chan error)
	return nil
}

func (s *Server) setupListener() (net.Listener, error) {
	port := strings.Split(s.addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %s: %w", port, err)
	}
	return lis, nil
}

func (s *Server) setupGRPCServer() *grpc.Server {
	grpcServer := grpc.NewServer()
	pb.RegisterGroupCacheServer(grpcServer, s)
	return grpcServer
}

func (s *Server) registerService(lis net.Listener, errChan chan error) {
	defer func() {
		if err := lis.Close(); err != nil {
			logger.LogrusObj.Errorf("failed to close listener: %v", err)
		}
	}()

	err := discovery.Register(serviceName, s.addr, s.stopSignal)
	if err != nil {
		logger.LogrusObj.Errorf("failed to register service: %v", err)
		errChan <- err
		return
	}

	// Wait for stop signal
	<-s.stopSignal
	logger.LogrusObj.Infof("service %s unregistered", s.addr)
}

func (s *Server) serveRequests(grpcServer *grpc.Server, lis net.Listener) error {
	// Serve blocks until the server stops or encounters an error
	if err := grpcServer.Serve(lis); err != nil && s.isRunning {
		return fmt.Errorf("failed to serve on %s: %w", s.addr, err)
	}
	return nil
}

// Stop gracefully shuts down the server and cleans up resources.
// It's safe to call Stop multiple times.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	// Signal service registration goroutine to stop
	close(s.stopSignal)

	// Update server state
	s.isRunning = false

	// Clean up resources
	s.cleanup()

	logger.LogrusObj.Infof("server %s stopped successfully", s.addr)

	return nil
}

func (s *Server) cleanup() {
	// Clear maps and help GC
	for k := range s.clients {
		delete(s.clients, k)
	}
	s.clients = nil
	s.consistHash = nil
}
