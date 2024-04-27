package etcd

import (
	"context"
	"time"

	"github.com/1055373165/ggcache/config"
	"github.com/1055373165/ggcache/utils/logger"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetPeers(service string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("create etcd client failed, err: %v", err)
		return []string{}, err
	}

	// etcdctl get "" --prefix 查询当前 etcd 中所有记录
	resp, err := cli.Get(ctx, service, clientv3.WithPrefix())
	if err != nil {
		logger.LogrusObj.Errorf("get peer addr list from etcd failed, err: %v", err)
		return []string{}, err
	}

	var peers []string
	for _, kv := range resp.Kvs {
		peers = append(peers, string(kv.Value))
	}

	logger.LogrusObj.Infof("get peer addr list from etcd success, peers: %v", peers)

	return peers, nil
}
