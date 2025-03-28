package server

import (
	"encoding/json"
	"log"
	"net/http"
	"radish/internal/cache"
)

type Server struct {
	cache *cache.Cache
}


func NewServer(c *cache.Cache) *Server {
	return &Server{cache: c}
}


func (s *Server) Start() {
	http.HandleFunc("/put", s.putHandler)
	http.HandleFunc("/get", s.getHandler)

	log.Println("Server running on port 7171...")
	log.Fatal(http.ListenAndServe(":7171", nil))
}


func (s *Server) putHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	json.NewDecoder(r.Body).Decode(&req)

	key, value := req["key"], req["value"]
	s.cache.Put(key, value)

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "OK",
		"message": "Key inserted/updated successfully.",
	})
}


func (s *Server) getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value, exists := s.cache.Get(key)

	if exists {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "OK",
			"key":    key,
			"value":  value,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ERROR",
			"message": "Key not found.",
		})
	}
}
