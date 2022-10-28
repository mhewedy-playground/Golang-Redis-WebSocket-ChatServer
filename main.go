package main

import (
	"chat/api"
	"chat/rdcon"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

const (
	// used to track users that used chat. mainly for listing users in the /users api, in real world chat app
	// such user list should be separated into user management module.
	usersKey       = "users"
	userChannelFmt = "user:%s:channels"
	ChannelsKey    = "channels"
)

func main() {

	rdcon.GetRedis().Client.SAdd(ChannelsKey, "general", "random")

	r := mux.NewRouter()

	r.Path("/ws/{username}").Methods("GET").HandlerFunc(api.ChatWebSocketHandler)
	r.Path("/user/{user}/channels").Methods("GET").HandlerFunc(api.UserChannelsHandler)
	r.Path("/users").Methods("GET").HandlerFunc(api.UsersHandler)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}
	fmt.Println("chat service started on port", port)
	log.Fatal(http.ListenAndServe(port, r))
}
