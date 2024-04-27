package main

import (
	"context"
	"fmt"
	"math/rand"

	pb "github.com/1055373165/ggcache/api/groupcachepb"
	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	etcdUrl     = "http://localhost:2379"
	serviceName = "GroupCache"
)

const ErrRPCCallNotFound = "rpc error: code = Unknown desc = record not found"

func main() {
	//bd := &ChihuoBuilder{addrs: map[string][]string{"/api": []string{"localhost:8001", "localhost:8002", "localhost:8003"}}}
	//resolver.Register(bd)
	config.InitConfig()
	etcdClient, err := clientv3.NewFromURL(etcdUrl)
	if err != nil {
		panic(err)
	}
	etcdResolver, err := resolver.NewBuilder(etcdClient)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.Dial(fmt.Sprintf("etcd:///%s", serviceName), grpc.WithResolvers(etcdResolver), grpc.WithTransportCredentials(insecure.NewCredentials()))
	logger.LogrusObj.Debugf("grpc dial connected success...")
	if err != nil {
		fmt.Printf("err: %v", err)
		return
	}

	ServerClient := pb.NewGroupCacheClient(conn)
	names := []string{"王五", "赵四", "李雷", "张三", "刘六", "陈七", "杨八", "吴九", "周十", "徐二",
		"孙明", "朱琪", "马华", "胡京", "郭士", "何东", "高北", "罗成", "林松", "赖林",
		"郑帅", "黄蓉", "韩梅", "顾桂", "汪松", "施云", "文希", "向荣", "梁宝", "宋江",
		"唐伯", "许利", "魏明", "蒋华", "沈丹", "韦石", "昌平", "苏波", "金山", "侯月",
		"邓光", "曹志", "彭波", "曾峰", "田野", "樊瑞", "程心", "袁思", "陆雨", "邹渊"}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	for i := 0; i < 100; i++ {
		names = append(names, fmt.Sprintf("%d", i))
	}
	rand.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})
	for {
		for _, name := range names {
			resp, err := ServerClient.Get(context.Background(), &pb.GetRequest{
				Group: "scores",
				Key:   name,
			})
			if err != nil {
				if ErrRPCCallNotFound != err.Error() {
					logger.LogrusObj.Fatalf("rpc call failed, err: %v", err)
					return
				} else {
					logger.LogrusObj.Warnf("没有查询到学生 %s 的成绩", name)
				}
			} else {
				logger.LogrusObj.Infof("rpc 调用成功, 学生 %s 的成绩为 %s", name, string(resp.Value))
			}
		}
	}
}
