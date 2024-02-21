package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/1055373165/Distributed_KV_Store/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// 从etcd中获取配置项（服务注册发现）
func GetPeers(prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cli, err := clientv3.New(DefaultEtcdConfig)
	if err != nil {
		logger.Logger.Error("create etcd client failed,err:", err)
		return []string{}, err
	}

	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Println("get peer addr list from etcd failed,err:", err)
		return []string{}, err
	}

	var peers []string
	for _, kv := range resp.Kvs {
		peers = append(peers, string(kv.Value))
	}

	logger.Logger.Info("get peer addr list from etcd success,peers:", peers)
	return peers, nil
}
