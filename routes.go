package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
)

var upgrader = websocket.Upgrader{} // use default options

func h(rdb *redis.Client, fn func(http.ResponseWriter, *http.Request, *redis.Client)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, rdb)
	}
}

type WSMsg struct {
	Content string `json:"content"`
	Type    int    `json:"type"`
	Command int    `json:"command"`
}

const (
	msgTypeChat = iota
	msgTypeCommand
)

const (
	commandSubscribe = iota
	commandUnsubscribe
)

func chatHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {

	r.Header.Set("Access-Control-Allow-Origin", "*")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		handleError(err, w)
		return
	}

	handleConnectionStarted(r, conn, rdb)
	handleConnectionEnded(r, conn, rdb)

	for {
		t, r, err := conn.NextReader()
		if err != nil {
			return
		}
		fmt.Println(t)

		nw, err := conn.NextWriter(t)
		if err != nil {
			handleError(err, w)
			break
		}
		if _, err := io.Copy(io.MultiWriter(nw, os.Stdout), r); err != nil {
			handleError(err, w)
			break
		}

		if err := nw.Close(); err != nil {
			handleError(err, w)
			break
		}
	}
}

func handleConnectionStarted(r *http.Request, conn *websocket.Conn, rdb *redis.Client) {
	username := r.URL.Query()["username"][0]
	fmt.Println("Connected", conn.RemoteAddr(), username)
	// TODO connect user here ... and put in user's list
}

func handleConnectionEnded(r *http.Request, conn *websocket.Conn, rdb *redis.Client) {
	username := r.URL.Query()["username"][0]
	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("connection closed for use", username)
		// TODO disconnect user here and remove it from users list
		return nil
	})
}

func channelsHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	// TODO list public channels
	/*	username := mux.Vars(r)["user"]

		if err := newUser(username).connect(rdb); err != nil {
			handleError(err, w)
			return
		}*/
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
