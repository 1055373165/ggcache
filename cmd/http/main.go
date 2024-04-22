package main

import (
	"flag"
	http_service "ggcache/internal/server/http"
)

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()
	http_service.Start(port, api)
}
