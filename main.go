package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var rdb *redis.Client

func main() {

	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	r := mux.NewRouter()

	r.Path("/chat").Methods("GET").HandlerFunc(h(rdb, chatHandler))
	r.Path("/channels").Methods("GET").HandlerFunc(h(rdb, channelsHandler))

	log.Fatal(http.ListenAndServe(":8080", r))
}
