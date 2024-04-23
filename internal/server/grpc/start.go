package grpc

import (
	"fmt"
	pb "ggcache/api"
	"ggcache/internal/middleware/etcd"
	"ggcache/internal/middleware/logger"
	"net"
	"strings"

	"google.golang.org/grpc"
)

/*
------------start service----------------
1. set server running status
2. initialize stop channel to notify registry stop keepalive lease
3. initialize tcp socket and start listening
4. register the customize rpc service to grpc, so that grpc can distribute the request to the server for processing
5. use etcd service registry which can directly obtain the client connection to the given service through the service name, that is, the grpc channel, and then create a client Stub, which implements the same method as the server and calls it directly
*/
func (s *Server) Start() {
	s.mu.Lock()
	if s.Status {
		s.mu.Unlock()
		fmt.Printf("server %s is already started\n", s.Addr)
		return
	}

	s.Status = true
	s.stopSignal = make(chan error)

	// waiting for client to connect
	port := strings.Split(s.Addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("failed to listen %s, error: %v\n", s.Addr, err)
		return
	}

	// getting a customized implementation of a grpc Server instance
	grpcServer := grpc.NewServer()
	// registers the service and its implementation to a concrete type that implements GroupCacheServer interface.
	pb.RegisterGGCacheServer(grpcServer, s)

	// service registry
	go func() {
		// register never return unless stop signal received (blocked)
		// for-select combination implement permanent blocking
		err := etcd.Register("ggcache", s.Addr, s.stopSignal)
		if err != nil {
			logger.Logger.Error(err.Error())
		}

		close(s.stopSignal)

		err = lis.Close()
		if err != nil {
			logger.Logger.Error(err.Error())
		}
		logger.Logger.Warnf("[%s] Revoke service and close tcp socket", s.Addr)
	}()

	logger.Logger.Infof("[%s] register service success\n", s.Addr)
	s.mu.Unlock()

	// Serve accepts incoming connections on the listener list, creating a new Server Transport and service Goroutine for each connection.
	// The service goroutines read gRPC requests and then call the registered handlers to reply to them.
	if err := grpcServer.Serve(lis); s.Status && err != nil {
		fmt.Printf("failed to serve %s, error: %v\n", s.Addr, err)
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
	s.stopSignal <- nil
	// change server running status
	s.Status = false
	s.fetchers = nil // help GC
	s.consistenthash = nil
}
