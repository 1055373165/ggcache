package http

import (
	"ggcache/internal/middleware/logger"
	"ggcache/internal/service"
	"log"
	"net/http"
)

func StartHTTPCacheServer(addr string, addrs []string, ggcache *service.Group) {
	peers := NewHTTPPool(addr)
	peers.UpdatePeers(addrs...)
	ggcache.RegisterPickerForGroup(peers)
	logger.Logger.Infof("service is running at %v", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func StartHTTPAPIServer(apiAddr string, ggcache *service.Group) {
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
	logger.Logger.Infof("fontend server is running at %v", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
