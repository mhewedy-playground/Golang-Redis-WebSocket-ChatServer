package main

import (
	"chat/api"
	"chat/user"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var rdb *redis.Client

func init() {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()
	rdb.SAdd(user.ChannelsKey, "general", "random")
}

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

	r.Path("/chat").Methods("GET").HandlerFunc(api.H(rdb, api.ChatHandler))
	r.Path("/user/{user}/channels").Methods("GET").HandlerFunc(api.H(rdb, api.UserChannelsHandler))
	r.Path("/users").Methods("GET").HandlerFunc(api.H(rdb, api.UsersHandler))

	log.Fatal(http.ListenAndServe(":8080", r))
}

func cleanup() {
	fmt.Println("Handle graceful shutdown ...")
	l := api.DisconnectUsers(rdb)
	fmt.Println(l, "users disconnected.")
	os.Exit(0)
}
