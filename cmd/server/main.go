package main

import (
	"radish/internal/cache"
	"radish/internal/server"
)

func main() {
	kvStore := cache.NewCache()
	srv := server.NewServer(kvStore)
	srv.Start()
}
