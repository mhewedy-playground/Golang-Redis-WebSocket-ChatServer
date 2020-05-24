package routes

import (
	"chat/user"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
)

var upgrader = websocket.Upgrader{} // use default options

var connectedUsers = make(map[string]*user.User)

func H(rdb *redis.Client, fn func(http.ResponseWriter, *http.Request, *redis.Client)) func(http.ResponseWriter, *http.Request) {
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

func ChatHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		handleError(err, w)
		return
	}

	err = onConnect(r, conn, rdb)
	if err != nil {
		handleError(err, w)
		return
	}
	onDisconnect(r, conn, rdb)

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

func onConnect(r *http.Request, conn *websocket.Conn, rdb *redis.Client) error {
	username := r.URL.Query()["username"][0]
	fmt.Println("Connected", conn.RemoteAddr(), username)

	u, err := user.Connect(rdb, username)
	if err != nil {
		return nil
	}
	connectedUsers[username] = u
	return nil
}

func onDisconnect(r *http.Request, conn *websocket.Conn, rdb *redis.Client) {
	username := r.URL.Query()["username"][0]

	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("connection closed for use", username)

		u := connectedUsers[username]
		if err := u.Disconnect(rdb); err != nil {
			return err
		}
		delete(connectedUsers, username)
		return nil
	})
}

func DisconnectUsers(rdb *redis.Client) int {
	l := len(connectedUsers)
	for _, u := range connectedUsers {
		_ = u.Disconnect(rdb)
	}
	connectedUsers = map[string]*user.User{}
	return l
}

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
