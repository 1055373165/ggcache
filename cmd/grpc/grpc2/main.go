package main

import (
	"flag"
	"fmt"
	"ggcache/config"
	"log"

	"ggcache/internal/middleware/etcd"
	"ggcache/internal/pkg/student/dao"
	"ggcache/internal/service"
)

var (
	port = flag.Int("port", 10000, "port")
	m    = map[int]string{
		10000: "localhost:10000",
		10001: "localhost:10001",
		10002: "localhost:10002",
	}
)

// 运行之前先运行 etcd/server_register_to_etcd 三个 put 将服务地址打入 etcd
func main() {
	config.InitConfig()
	dao.InitDB()
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", m[*port], *port)
	groupManager := service.NewGroupManager([]string{"scores", "student"}, addr)

	svr, err := service.NewServer(addr)
	if err != nil {
		panic(err)
	}

	// put clusters/nodeadress nodeadress
	addrs, err := etcd.GetPeers("clusters")
	if err != nil {
		addrs = []string{"localhost:10000"}
	}

	// Place the node on the hash ring
	svr.SetPeers(addrs)
	// Register the service Picker for Group
	groupManager["scores"].RegisterServer(svr)
	log.Println("group scores is running at ", addr)

	// Start the service (register the service to etcd, calculate consistent hash)
	// Our Service Name is groupcache
	err = svr.Start()
	if err != nil {
		log.Fatal(err)
	}
}
