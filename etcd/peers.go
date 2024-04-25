package etcd

import (
	"context"
	"time"

	"github.com/1055373165/Distributed_KV_Store/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Use the etcd prefix query to get the addresses of all service instances
func GetPeers(prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cli, err := clientv3.New(DefaultEtcdConfig)
	if err != nil {
		logger.Logger.Errorf("create etcd client failed, err: %v", err)
		return []string{}, err
	}

	// etcdctl get "" --prefix 查询当前 etcd 中所有记录
	resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Logger.Errorf("get peer addr list from etcd failed, err: %v", err)
		return []string{}, err
	}

	var peers []string
	for _, kv := range resp.Kvs {
		peers = append(peers, string(kv.Value))
	}

	logger.Logger.Infof("get peer addr list from etcd success, peers: %v", peers)

	return peers, nil
}
