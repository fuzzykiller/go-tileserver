package main

import (
	"gopkg.in/tylerb/graceful.v1"
	"time"
	"net/http"
)

func main() {
	RegisterTileDatabase("/tmp/bla.db", "")
	mux := http.NewServeMux()
	mux.HandleFunc("/", GetTile)
	graceful.Run("127.0.0.1:8080", 1 * time.Second, mux)
}
