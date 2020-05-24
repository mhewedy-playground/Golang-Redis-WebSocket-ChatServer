package api

import (
	"chat/user"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"net/http"
)

func ChannelsHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	// TODO list public channels
	/*	username := mux.Vars(r)["user"]

		if err := newUser(username).connect(rdb); err != nil {
			handleError(err, w)
			return
		}*/
}

func UsersHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {

	list, err := user.List(rdb)
	if err != nil {
		handleError(err, w)
		return
	}
	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		handleError(err, w)
		return
	}
}

/*
func subscribeHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	username := mux.Vars(r)["user"]
	channel := mux.Vars(r)["channel"]

	if err := newUser(username).subscribe(rdb, channel); err != nil {
		handleError(err, w)
		return
	}
}

func unsubscribeHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	username := mux.Vars(r)["user"]
	channel := mux.Vars(r)["channel"]

	if err := newUser(username).unsubscribe(rdb, channel); err != nil {
		handleError(err, w)
		return
	}
}*/

func handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
}
