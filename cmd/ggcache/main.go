package main

import (
	"flag"
	"log"

	"github.com/1055373165/ggcache/internal/app"
	"github.com/1055373165/ggcache/configs"
)

var (
	configPath  = flag.String("config", "", "config file path")
	port        = flag.Int("port", 0, "service node port (override config)")
	metricsPort = flag.Int("metricsPort", 0, "metrics port (override config)")
	pprofPort   = flag.Int("pprofPort", 0, "pprof port (override config)")
)

// 封装全部启动逻辑（加载配置、DB 初始化、pprof/metrics、etcd 发现、gRPC 服务启动）
func main() {
	flag.Parse()

	if err := configs.InitConfig(*configPath); err != nil {
		log.Fatal(err)
	}

	if err := app.Run(*port, *metricsPort, *pprofPort); err != nil {
		log.Fatal(err)
	}
}
