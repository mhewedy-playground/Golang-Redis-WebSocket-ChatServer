package main

import (
	"chat/routes"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var rdb *redis.Client

/**
flushall
sadd user:wael:channels news general
sadd user:mazen:channels news general
*/
func main() {

	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cleanup()
		}
	}()

	r := mux.NewRouter()

	r.Path("/chat").Methods("GET").HandlerFunc(routes.H(rdb, routes.ChatHandler))
	r.Path("/channels").Methods("GET").HandlerFunc(routes.H(rdb, routes.ChannelsHandler))

	log.Fatal(http.ListenAndServe(":8080", r))
}

func cleanup() {
	fmt.Println("cleaning ...")
	routes.DisconnectUsers(rdb)
	os.Exit(0)
}
