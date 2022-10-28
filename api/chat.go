package api

import (
	"chat/user"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader websocket.Upgrader

var connectedUsers = make(map[string]*user.User)

type msg struct {
	Content string `json:"content,omitempty"`
	Channel string `json:"channel,omitempty"`
	Command int    `json:"command,omitempty"`
	Err     string `json:"err,omitempty"`
}

const (
	commandSubscribe = iota
	commandUnsubscribe
	commandChat
)

func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		handleWSError(err, conn)
		return
	}

	err = onConnect(r, conn)
	if err != nil {
		handleWSError(err, conn)
		return
	}

	closeCh := onDisconnect(r, conn)

	onChannelMessage(conn, r)

loop:
	for {
		select {
		case <-closeCh:
			break loop
		default:
			onUserMessage(conn, r)
		}
	}
}

func onConnect(r *http.Request, conn *websocket.Conn) error {
	//username := r.URL.Query()["username"][0]
	username := mux.Vars(r)["username"]
	log.Println("connected from:", conn.RemoteAddr(), "user:", username)

	u, err := user.Connect(username)
	if err != nil {
		return err
	}
	connectedUsers[username] = u
	return nil
}

func onDisconnect(r *http.Request, conn *websocket.Conn) chan struct{} {

	closeCh := make(chan struct{})

	//username := r.URL.Query()["username"][0]
	username := mux.Vars(r)["username"]
	conn.SetCloseHandler(func(code int, text string) error {
		log.Println("connection closed for user", username)

		u := connectedUsers[username]
		if err := u.Disconnect(username); err != nil {
			return err
		}
		delete(connectedUsers, username)
		close(closeCh)
		return nil
	})

	return closeCh
}

func onUserMessage(conn *websocket.Conn, r *http.Request) {
	var msg msg

	if err := conn.ReadJSON(&msg); err != nil {
		handleWSError(err, conn)
		return
	}

	//username := r.URL.Query()["username"][0]
	username := mux.Vars(r)["username"]
	u := connectedUsers[username]

	switch msg.Command {
	case commandSubscribe:
		if err := u.Subscribe(msg.Channel); err != nil {
			handleWSError(err, conn)
		}
	case commandUnsubscribe:
		if err := u.Unsubscribe(msg.Channel); err != nil {
			handleWSError(err, conn)
		}
	case commandChat:
		if err := user.Chat(msg.Channel, msg.Content); err != nil {
			handleWSError(err, conn)
		}
	}
}

func onChannelMessage(conn *websocket.Conn, r *http.Request) {
	//username := r.URL.Query()["username"][0]
	username := mux.Vars(r)["username"]
	log.Printf(username + " OnChannelMessage \n")
	log.Printf("online users length [%d] \n", len(connectedUsers))

	for us := range connectedUsers {
		log.Println(us)
	}

	u := connectedUsers[username]

	go func() {
		for m := range u.MessageChan {
			log.Printf("On Channel Message %s, %s, %s\n", u.Name, m.Payload, m.Channel)

			msg := msg{
				Content: m.Payload,
				Channel: m.Channel,
			}

			if err := conn.WriteJSON(msg); err != nil {
				fmt.Println(err)
			}
		}
	}()
}

func handleWSError(err error, conn *websocket.Conn) {

	if conn != nil {
		if err := conn.WriteJSON(msg{Err: err.Error()}); err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("Websocket Connection is nil")
	}

}
