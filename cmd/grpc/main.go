package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/1055373165/ggcache/config"
	etcdservice "github.com/1055373165/ggcache/internal/middleware/etcd"
	dao "github.com/1055373165/ggcache/internal/pkg/student/dao"
	"github.com/1055373165/ggcache/internal/service"
	"github.com/1055373165/ggcache/utils/logger"
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

	groupManager := service.NewGroupManager([]string{"scores"}, addr)

	updateChan := make(chan bool)
	svr, err := service.NewServer(updateChan, addr)
	if err != nil {
		panic(err)
	}

	// put clusters/nodeAddress nodeAddress
	addrs, err := etcdservice.GetPeers("clusters")
	if err != nil {
		addrs = []string{"localhost:9999"}
	}
	logger.LogrusObj.Infof("获取到用于设置一致性哈希环的地址列表: %v", addrs)
	// Place the node on the hash ring
	svr.SetPeers(addrs)

	// Register the service Picker for Group
	groupManager["scores"].RegisterServer(svr)
	logger.LogrusObj.Debug("groupcache is running at ", addr)

	// Start the service (register the service to etcd, calculate consistent hash)
	// Our Service Name is groupcache
	svr.Start()
	if err != nil {
		log.Fatal(err)
	}
}
