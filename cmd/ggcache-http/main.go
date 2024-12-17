package main

import (
	"flag"
	"fmt"

	"github.com/1055373165/ggcache/internal/cache"
)

var (
	port = flag.Int("port", 9999, "service node default port")
	api  = flag.Bool("api", false, "Start a api server?")
	// assuming that the api server and http://127.0.0.1:8001 are running on the same physical machine
	apiServerAddr1 = "http://127.0.0.1:8000"
	apiServerAddr2 = "http://127.0.0.1:8001"
)

func main() {
	flag.Parse()

	/* if you have a configuration center, both api client and http server configurations can be pulled from the configuration center */
	serverAddrMap := map[int]string{
		9999:   "http://127.0.0.1:9999",
		10000:  "http://127.0.0.1:10000",
		100001: "http://127.0.0.1:100001",
	}
	var serverAddrs []string
	for _, v := range serverAddrMap {
		serverAddrs = append(serverAddrs, v)
	}

	gm := cache.NewGroupManager([]string{"scores", "website"}, fmt.Sprintf("127.0.0.1:%d", *port))
	//  start http api server for client load balancing
	if *api {
		go cache.StartHTTPAPIServer(apiServerAddr1, gm["scores"])
		go cache.StartHTTPAPIServer(apiServerAddr2, gm["website"])
	}
	// start http server to provide caching service
	cache.StartHTTPCacheServer(serverAddrMap[*port], []string(serverAddrs), gm["scores"])
	cache.StartHTTPCacheServer(serverAddrMap[*port], []string(serverAddrs), gm["website"])
}