package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/internal/middleware/etcd"
	"github.com/1055373165/ggcache/internal/pkg/student/dao"
	"github.com/1055373165/ggcache/internal/service"
)

var (
	port = flag.Int("port", 9999, "port")
)

// 运行之前先运行 etcd/server_register_to_etcd 三个 put 将服务地址打入 etcd
func main() {
	config.InitConfig()
	dao.InitDB()
	flag.Parse()

	addr := fmt.Sprintf("localhost:%d", *port)
	groupManager := service.NewGroupManager([]string{"scores", "student"}, addr)
	svr, err := service.NewServer(addr)
	if err != nil {
		panic(err)
	}

	// put clusters/nodeadress nodeadress
	addrs, err := etcd.GetPeers("clusters")
	if err != nil {
		panic(err)
	}
	if len(addrs) == 0 {
		addrs = []string{"localhost:9999"}
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
