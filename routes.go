package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"net/http"
)

func h(rdb *redis.Client, fn func(http.ResponseWriter, *http.Request, *redis.Client)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, rdb)
	}
}

func connectHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	username := mux.Vars(r)["user"]

	if err := newUser(username).connect(rdb); err != nil {
		handelError(err, w)
		return
	}
}

func subscribeHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	username := mux.Vars(r)["user"]
	channel := mux.Vars(r)["channel"]

	if err := newUser(username).subscribe(rdb, channel); err != nil {
		handelError(err, w)
		return
	}
}

func unsubscribeHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	username := mux.Vars(r)["user"]
	channel := mux.Vars(r)["channel"]

	if err := newUser(username).unsubscribe(rdb, channel); err != nil {
		handelError(err, w)
		return
	}
}

func handelError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
}
