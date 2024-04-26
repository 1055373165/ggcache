package main

import (
	"context"
	"fmt"
	"time"

	"ggcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	//初始化
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		fmt.Println("new clientv3 failed,err:", err)
		return
	}

	fmt.Println("connect to etcd success!")
	defer cli.Close()

	//put
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = cli.Put(ctx, "GroupCache", "localhost:10000")
	if err != nil {
		logger.LogrusObj.Error("put groupcache service to etcd failed")
		return
	}

	fmt.Println("put groupcache service to etcd success!")
}
