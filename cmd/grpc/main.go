package main

import (
	"flag"
	"fmt"
	conf "ggcache/configs"
	etcdservice "ggcache/internal/middleware/etcd"
	grpcserver "ggcache/internal/server/grpc"
	"ggcache/internal/service"
)

var (
	port        = flag.Int("port", 9999, "service node port")
	serviceName = "ggcache"
)

func main() {
	conf.Init()
	flag.Parse()

	ggcache := service.NewGroupInstance("scores")
	// grpc node local service address
	serviceAddr := fmt.Sprintf("localhost:%d", *port)

	// get a grpc service instance
	ch := make(chan bool)
	svr, err := grpcserver.NewServer(ch, serviceAddr)
	if err != nil {
		fmt.Printf("acquire grpc server instance failed, %v\n", err)
		return
	}

	go etcdservice.DynamicServices(ch, serviceName)
	// Server implemented Pick interface, register a node selector for ggcache
	svr.UpdatePeers(etcdservice.ListServicePeers(serviceName))
	ggcache.RegisterPickerForGroup(svr)

	// start grpc service
	svr.Start()
}
