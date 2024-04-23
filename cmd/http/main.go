package main

import (
	"flag"
	conf "ggcache/configs"
	httpserver "ggcache/internal/server/http"
	"ggcache/internal/service"
)

var (
	port = flag.Int("port", 8001, "service node default port")
	api  = flag.Bool("api", false, "Start a api server?")
	// assuming that the api server and http://127.0.0.1:8001 are running on the same physical machine
	apiServerAddr = "http://127.0.0.1:9999"
)

func main() {
	// proposal: you can use viper
	conf.Init()
	flag.Parse()

	/* if you have a configuration center, both api client and http server configurations can be pulled from the configuration center */
	serverAddrMap := map[int]string{
		8001: "http://127.0.0.1:8001",
		8002: "http://127.0.0.1:8002",
		8003: "http://127.0.0.1:8003",
	}
	var serverAddrs []string
	for _, v := range serverAddrMap {
		serverAddrs = append(serverAddrs, v)
	}

	ggcache := service.NewGroupInstance("scores")
	//  start http api server for client load balancing
	if *api {
		go httpserver.StartHTTPAPIServer(apiServerAddr, ggcache)
	}
	// start http server to provide caching service
	httpserver.StartHTTPCacheServer(serverAddrMap[*port], []string(serverAddrs), ggcache)
}
