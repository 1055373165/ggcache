package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

var client *clientv3.Client

// 不同 peer 节点的地址信息
type PeerEntry struct {
	addr string
}

// 初始化 etcd 客户端
func Init(address []string, timeout time.Duration) error {
	var err error
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: timeout,
	})
	return err
}

// 从 etcd 中拉取 peer 节点地址信息
func GetPeersAddr(key string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// key_addr 规定 client 注册时默认以服务器名开头 /service/ip 后面跟上自己的访问地址
	// 搜索所有以 /service/ 开头的键
	prefix := "/service/"
	// 查询etcd中以指定前缀开头的键值对
	resp, err := client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// 解析查询结果，获取服务地址
	var addrs []string
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}

	return addrs, nil
}

const etcdUrl = "http://localhost:2379"
const serviceName = "groupcache"
const ttl = 10

var etcdClient *clientv3.Client

func etcdRegister(addr string) error {
	log.Printf("etcdRegister %s\b", addr)
	etcdClient, err := clientv3.NewFromURL(etcdUrl)

	if err != nil {
		return err
	}

	em, err := endpoints.NewManager(etcdClient, serviceName)
	if err != nil {
		return err
	}

	lease, _ := etcdClient.Grant(context.TODO(), ttl)

	err = em.AddEndpoint(context.TODO(), fmt.Sprintf("%s/%s", serviceName, addr), endpoints.Endpoint{Addr: addr}, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}
	//etcdClient.KeepAlive(context.TODO(), lease.ID)
	alive, err := etcdClient.KeepAlive(context.TODO(), lease.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			<-alive
			fmt.Println("etcd server keep alive")
		}
	}()

	return nil
}

func etcdUnRegister(addr string) error {
	log.Printf("etcdUnRegister %s\b", addr)
	if etcdClient != nil {
		em, err := endpoints.NewManager(etcdClient, serviceName)
		if err != nil {
			return err
		}
		err = em.DeleteEndpoint(context.TODO(), fmt.Sprintf("%s/%s", serviceName, addr))
		if err != nil {
			return err
		}
		return err
	}

	return nil
}
