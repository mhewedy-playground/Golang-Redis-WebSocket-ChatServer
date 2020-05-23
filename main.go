package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"os"
)

type user struct {
	name  string
	rooms []string

	stopListener    chan bool
	listenerRunning bool

	roomsHandler map[string]*redis.PubSub
}

func newUser(name string) *user {
	return &user{
		name:         name,
		stopListener: make(chan bool),
		roomsHandler: make(map[string]*redis.PubSub),
	}
}

func (u *user) connect(rdb *redis.Client) error {
	// get all user rooms (from DB) and start subscribe
	r, err := rdb.SMembers(fmt.Sprintf("user:%s:rooms", u.name)).Result()
	if err != nil {
		return err
	}
	if len(r) == 0 {
		return nil
	}

	// if use has saved rooms on server, then subscribe on each room
	for _, room := range r {
		return u.subscribe(room, rdb)
	}

	return nil
}

func (u *user) subscribe(room string, rdb *redis.Client) error {
	// check if already subscribed
	for i := range u.rooms {
		if u.rooms[i] == room {
			return nil
		}
	}

	// save user room to server
	userRoomsKey := fmt.Sprintf("user:%s:rooms", u.name)
	if err := rdb.SAdd(userRoomsKey, room).Err(); err != nil {
		return err
	}

	// get all user rooms from server, set it as user.rooms and start subscribing
	r, err := rdb.SMembers(userRoomsKey).Result()
	if err != nil {
		return err
	}
	u.rooms = r

	if u.listenerRunning {
		u.stopListener <- true
	}

	u.doSubscribe(room, rdb)

	return nil
}

func (u *user) doSubscribe(room string, rdb *redis.Client) {
	// subscribe all rooms in one request
	pubSub := rdb.Subscribe(u.rooms...)
	// keep room handler to be used in unsubscribe
	u.roomsHandler[room] = pubSub

	// The Listener
	go func() {
		u.listenerRunning = true
		fmt.Println("starting the listener for user:", u.name, "on rooms:", u.rooms)
		for {
			select {
			case msg, ok := <-pubSub.Channel():
				if !ok {
					break
				}
				fmt.Println("msg:", msg.Payload, "room:", msg.Channel)

			case <-u.stopListener:
				fmt.Println("Stop listening for user:", u.name, "on old rooms")

				for k, v := range u.roomsHandler {
					if err := v.Unsubscribe(); err != nil {
						fmt.Fprintln(os.Stderr, "unable to unsubscribe", err)
					}
					delete(u.roomsHandler, k)
				}
				break
			}
		}
	}()
}

func (u *user) unsubscribe(room string, rdb *redis.Client) error {
	return nil
}

var rdb *redis.Client

func main() {

	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	u := newUser("Wael")
	if err := u.connect(rdb); err != nil {
		log.Fatal(err)
	}

	// TODO:
	/*
		1. save 2 or 3 rooms under "default_rooms" list on server
		on on each connect, let the user subscribe to such "default_rooms"

		2. allow user to create a new room on demand (he can invite other users by room name)
		(make it room:user:<user supplied name>) might make it visible with "default_rooms" as well

		3. allow direct chat by create a room with name "room:direct:user1:user2" make both users subscribe to the room

		4. create a list room api to list all public rooms

	*/

	//
	//if err := u.subscribe("general", rdb); err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := u.subscribe("programming", rdb); err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := u.subscribe("New", rdb); err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := u.subscribe("Old", rdb); err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := u.subscribe("OldPlusPlus", rdb); err != nil {
	//	log.Fatal(err)
	//}

	/*
		r := mux.NewRouter()

		r.Path("/recipe").Methods("POST").HandlerFunc(createHandler)
		r.Path("/recipe/{id}").Methods("PUT").HandlerFunc(updateHandler)
		r.Path("/recipe/{id}").Methods("GET").HandlerFunc(getHandler)
		r.Path("/recipes").Methods("GET").HandlerFunc(listHandler)

		log.Fatal(http.ListenAndServe(":8080", r))
	*/

	select {}
}
