package http

import (
	"errors"
	conf "ggcache/configs"
	"ggcache/internal/middleware/logger"
	db "ggcache/internal/middleware/mysql"
	"ggcache/internal/service"
	"log"
	"net/http"
)

func Start(port int, api bool) {
	conf.Init()
	// http client
	apiAddr := "http://localhost:9999"
	// http server set
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	ggcache := NewGroupInstance("scores")
	if api {
		go startAPIServer(apiAddr, ggcache)
	}
	startCacheServer(addrMap[port], []string(addrs), ggcache)
}

func NewGroupInstance(groupName string) *service.Group {
	g := service.NewGroup(groupName, "lru", 1<<10, service.RetrieveFunc(func(key string) ([]byte, error) {
		// 从后端数据库中查找
		logger.Logger.Info("进入 RetrieveFunc, 数据库中查询....")

		var scores []*db.Student
		db.DB.Where("name = ?", key).Find(&scores)
		if len(scores) == 0 {
			logger.Logger.Info("后端数据库中也查询不到...")
			return []byte{}, errors.New("record not found")
		}

		logger.Logger.Infof("成功从后端数据库中查询到学生 %s 的分数：%s", key, scores[0].Score)
		return []byte(scores[0].Score), nil
	}))
	InitDataWithGroup()
	return g
}

func startCacheServer(addr string, addrs []string, ggcache *service.Group) {
	peers := NewHTTPPool(addr)
	peers.UpdatePeers(addrs...)
	ggcache.RegisterPickerForGroup(peers)
	log.Println("service is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, ggcache *service.Group) {
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
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func InitDataWithGroup() {
	// 先往数据库中存一些数据（慢速数据库）
	db.DB.Create(&db.Student{Name: "张三", Score: "10"})
	db.DB.Create(&db.Student{Name: "李四", Score: "100"})
	db.DB.Create(&db.Student{Name: "王五", Score: "1000"})
	db.DB.Create(&db.Student{Name: "李六", Score: "10000"})
	db.DB.Create(&db.Student{Name: "赵七", Score: "100000"})
	db.DB.Create(&db.Student{Name: "孙八", Score: "1000000"})
	db.DB.Create(&db.Student{Name: "钱九", Score: "10000000"})
	db.DB.Create(&db.Student{Name: "周十", Score: "100000000"})
	db.DB.Create(&db.Student{Name: "one", Score: "10"})
	db.DB.Create(&db.Student{Name: "two", Score: "100"})
	db.DB.Create(&db.Student{Name: "three", Score: "1000"})
	db.DB.Create(&db.Student{Name: "three", Score: "10000"})
	db.DB.Create(&db.Student{Name: "four", Score: "100000"})
	db.DB.Create(&db.Student{Name: "five", Score: "1000000"})
	db.DB.Create(&db.Student{Name: "six", Score: "10000000"})
	logger.Logger.Info("数据库数据插入成功...")
}
