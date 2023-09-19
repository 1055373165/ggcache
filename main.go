package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/1055373165/distributekv/conf"
	group "github.com/1055373165/distributekv/distributekv"
	services "github.com/1055373165/distributekv/etcd"
)

var (
	port = flag.Int("port", 9999, "port")
)

func main() {
	conf.Init()
	flag.Parse()
	// 新建 cache 实例
	g := group.NewGroupInstance("scores")

	// New 一个自己实现的服务实例
	addr := fmt.Sprintf("localhost:%d", *port)
	svr, err := group.NewServer(addr)
	if err != nil {
		log.Fatal(err)
	}

	// 设置同伴节点包括自己（同伴的地址从 etcd 中获取）
	addrs, err := services.GetPeers("clusters")
	if err != nil { // 如果查询失败使用默认的地址
		addrs = []string{"localhost:9999"}
	}
	fmt.Println("从 etcd 处获取的 server 地址", addrs)
	// 将节点打到哈希环上
	svr.SetPeers(addrs)
	// 为 Group 注册服务 Picker
	g.RegisterServer(svr)
	log.Println("groupcache is running at ", addr)

	// 启动服务（注册服务至 etcd、计算一致性 hash）
	err = svr.Start()
	if err != nil {
		log.Fatal(err)
	}
}
