package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/1055373165/Distributed_KV_Store/conf"
	group "github.com/1055373165/Distributed_KV_Store/distributekv"
	services "github.com/1055373165/Distributed_KV_Store/etcd"
)

var (
	port = flag.Int("port", 9999, "port")
)

// 运行之前先运行 etcd/server_register_to_etcd 三个 put 将服务地址打入 etcd
func main() {
	conf.Init()
	flag.Parse()

	g := group.NewGroupInstance("scores")

	addr := fmt.Sprintf("localhost:%d", *port)
	svr, err := group.NewServer(addr)
	if err != nil {
		panic(err)
	}

	// put clusters/nodeadress nodeadress
	addrs, err := services.GetPeers("clusters")
	if err != nil {
		addrs = []string{"localhost:9999"}
	}

	// Place the node on the hash ring
	svr.SetPeers(addrs)
	// Register the service Picker for Group
	g.RegisterServer(svr)
	log.Println("groupcache is running at ", addr)

	// Start the service (register the service to etcd, calculate consistent hash)
	// Our Service Name is groupcache
	err = svr.Start()
	if err != nil {
		log.Fatal(err)
	}
}
