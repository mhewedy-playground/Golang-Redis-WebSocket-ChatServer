package api

import (
	"chat/user"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{} // use default options

var connectedUsers = make(map[string]*user.User)

func H(rdb *redis.Client, fn func(http.ResponseWriter, *http.Request, *redis.Client)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, rdb)
	}
}

type msg struct {
	Content string `json:"content"`
	Channel string `json:"channel"`
	Command int    `json:"command,omitempty"`
}

const (
	commandSubscribe = iota
	commandUnsubscribe
	commandChat
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

	onReceiveMessage(conn, r, rdb)

	for {
		onSendMessage(conn, r, rdb)
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

func onSendMessage(conn *websocket.Conn, r *http.Request, rdb *redis.Client) {

	var msg msg

	err := conn.ReadJSON(&msg)
	if err != nil {
		fmt.Println("Error reading json.", err)
	}

	fmt.Printf("Got message: %#v\n", msg)

	username := r.URL.Query()["username"][0]
	_ = connectedUsers[username]

	//u.SendMessage()
}

func onReceiveMessage(conn *websocket.Conn, r *http.Request, rdb *redis.Client) {

	username := r.URL.Query()["username"][0]
	u := connectedUsers[username]

	go func() {
		for m := range u.MessageChan {

			msg := msg{
				Content: m.Payload,
				Channel: m.Channel,
				Command: 0,
			}

			if err := conn.WriteJSON(msg); err != nil {
				fmt.Println(err)
			}
		}

	}()
}

func DisconnectUsers(rdb *redis.Client) int {
	l := len(connectedUsers)
	for _, u := range connectedUsers {
		_ = u.Disconnect(rdb)
	}
	connectedUsers = map[string]*user.User{}
	return l
}
