package cache

import (
	"log"
	"net/http"

	"github.com/1055373165/ggcache/pkg/common/logger"
)

func StartHTTPCacheServer(addr string, addrs []string, ggcache *Group) {
	peers := NewHTTPPool(addr)
	peers.UpdatePeers(addrs...)
	ggcache.RegisterServer(peers)
	logger.LogrusObj.Infof("service is running at %v", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// todo: gin 路由拆分请求负载
func StartHTTPAPIServer(apiAddr string, ggcache *Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := ggcache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.Bytes())
		}))
	logger.LogrusObj.Infof("fontend server is running at %v", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
