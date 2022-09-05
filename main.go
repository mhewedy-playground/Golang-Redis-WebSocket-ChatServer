package main

import (
	"chat/api"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {

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
