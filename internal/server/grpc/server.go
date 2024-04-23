package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "ggcache/api"
	etcdservice "ggcache/internal/middleware/etcd"
	"ggcache/internal/middleware/logger"
	"ggcache/internal/service"
	"ggcache/internal/service/consistenthash"

	"ggcache/utils"
)

/*
grpc server provides communication capabilities between groupcache.
In this way, groupcache deployed on other machines can obtain the cache by accessing the server.
*/
const (
	defaultAddr     = "127.0.0.1:9999"
	defaultReplicas = 50
)

var _ service.Picker = (*Server)(nil)

// server and group are decoupled, so the server must implement concurrency control by itself
type Server struct {
	pb.UnimplementedGroupCacheServer

	Addr           string     // ip:port
	Status         bool       // running status
	stopSignal     chan error // registery revoke
	mu             sync.Mutex
	consistenthash *consistenthash.ConsistentHash
	fetchers       map[string]*grpcFetcher
	update         chan bool
}

func NewServer(update chan bool, addr string) (*Server, error) {
	if addr == "" {
		addr = defaultAddr
	}

	if !utils.ValidPeerAddr(addr) {
		return nil, fmt.Errorf("invalid addr %s, expect address format is x.x.x.x:port", addr)
	}

	return &Server{Addr: addr, update: update}, nil
}

// core method 1
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	// pb provides accessors for fields
	groupName, key := req.GetGroup(), req.GetKey()
	resp := &pb.GetResponse{}
	logger.Logger.Infof("[ggcache server %s] Recv RPC Request - (%s)/(%s)", s.Addr, groupName, key)

	if groupName == "" || key == "" {
		return resp, fmt.Errorf("key and group name is required parameters")
	}

	// get the group object corresponding to groupName from group manager
	g := service.GetGroup(groupName)
	if g == nil {
		return resp, fmt.Errorf("no such group %s", groupName)
	}

	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}

	// return read-only data
	resp.Value = view.Bytes()
	return resp, nil
}

// implement the Picker interface to gain the ability to locate service server(peer) based on key
func (s *Server) Pick(key string) (service.Fetcher, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peerAddress := s.consistenthash.GetTruthNode(key)
	if peerAddress == s.Addr {
		logger.Logger.Infof("oohhh! pick myself, I am %s", s.Addr)
		// upper layer get the value of the key locally after receiving false
		return nil, false
	}

	logger.Logger.Infof("[peer %s receive rpc request] pick remote peer: %s", s.Addr, peerAddress)
	return s.fetchers[peerAddress], true
}

// implementing load balancing using consistent hashing
func (s *Server) UpdatePeers(peerAddrs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(peerAddrs) == 0 {
		peerAddrs = []string{"127.0.0.1:9999"}
	}

	s.consistenthash = consistenthash.NewConsistentHash(defaultReplicas, nil)
	s.consistenthash.AddTruthNode(peerAddrs)
	s.fetchers = make(map[string]*grpcFetcher)

	for _, peerAddr := range peerAddrs {
		if !utils.ValidPeerAddr(peerAddr) {
			panic(fmt.Sprintf("[peer %s] invalid address format, expect x.x.x.x:port", peerAddr))
		}

		// demo: ggcache/127.0.0.1:9999
		// both service registry and service discovery processes need to follow the same format
		serviceNameWithPrefix := fmt.Sprintf("ggcache/%s", peerAddr)
		s.fetchers[peerAddr] = NewGrpcFetcher(serviceNameWithPrefix)
	}

	logger.Logger.Infof("UpdatePeers, peers: %v", peerAddrs)

	go func() {
		for {
			select {
			case <-s.update:
				go s.reconstruct()
			case <-s.stopSignal:
				s.Stop()
			default:
				time.Sleep(time.Millisecond * 500)
			}
		}
	}()
}

func (s *Server) reconstruct() {
	serviceList := etcdservice.ListServicePeers("ggcache")
	if len(serviceList) == 0 {
		return
	}

	s.consistenthash = consistenthash.NewConsistentHash(defaultReplicas, nil)
	s.consistenthash.AddTruthNode(serviceList)
	s.fetchers = make(map[string]*grpcFetcher)

	for _, peerAddr := range serviceList {
		if !utils.ValidPeerAddr(peerAddr) {
			panic(fmt.Sprintf("[peer %s] invalid address format, expect x.x.x.x:port", peerAddr))
		}

		// demo: ggcache/127.0.0.1:9999
		// both service registry and service discovery processes need to follow the same format
		serviceNameWithPrefix := fmt.Sprintf("ggcache/%s", peerAddr)
		s.fetchers[peerAddr] = NewGrpcFetcher(serviceNameWithPrefix)
	}
	logger.Logger.Infof("hash ring reconstruct, contain service peer %v", serviceList)
}
