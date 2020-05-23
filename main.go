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

	r.Path("/user/{user}/connect").Methods("POST").HandlerFunc(h(rdb, connectHandler))
	r.Path("/user/{user}/subscribe/{channel}").Methods("POST").HandlerFunc(h(rdb, subscribeHandler))
	r.Path("/user/{user}/unsubscribe/{channel}").Methods("POST").HandlerFunc(h(rdb, unsubscribeHandler))

	log.Fatal(http.ListenAndServe(":8080", r))
}
