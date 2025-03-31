package main

import (
	"radish/internal/cache"
	"radish/internal/server"
)

func main() {
	

	kvStore := cache.NewCacheManager(16)
	srv := server.NewTCPServer(kvStore)
	srv.Start()
}
