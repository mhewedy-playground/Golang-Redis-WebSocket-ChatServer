package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
)

type user struct {
	name        string
	rooms       []string
	stopRunning chan bool
	running     bool
	roomsPubsub map[string]*redis.PubSub
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

	for _, room := range u.rooms {
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

	fmt.Println("Cont", room)

	// add room to user
	userRooms := fmt.Sprintf("user:%s:rooms", u.name)
	if err := rdb.SAdd(userRooms, room).Err(); err != nil {
		return err
	}

	// get all user rooms (from DB) and start subscribe
	r, err := rdb.SMembers(userRooms).Result()
	if err != nil {
		return err
	}
	u.rooms = r

	if u.running {
		u.stopRunning <- true
	}

	u.doSubscribe(room, rdb)

	return nil
}

func (u *user) doSubscribe(room string, rdb *redis.Client) {
	pubSub := rdb.Subscribe(u.rooms...)
	u.roomsPubsub[room] = pubSub

	go func() {
		u.running = true
		fmt.Println("starting the listener for user:", u.name, "on rooms:", u.rooms)
		for {
			select {

			case msg, ok := <-pubSub.Channel():
				if !ok {
					break
				}
				fmt.Println(msg.Payload, msg.Channel)

			case <-u.stopRunning:
				fmt.Println("Stop listening for user:", u.name, "on old rooms")

				for k, v := range u.roomsPubsub {
					if err := v.Unsubscribe(); err != nil {
						fmt.Println("unable to unsubscribe", err)
					}
					delete(u.roomsPubsub, k)
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

	u := &user{
		name:        "Wael",
		stopRunning: make(chan bool),
		roomsPubsub: make(map[string]*redis.PubSub),
	}

	if err := u.connect(rdb); err != nil {
		log.Fatal(err)
	}

	if err := u.subscribe("general", rdb); err != nil {
		log.Fatal(err)
	}

	if err := u.subscribe("programming", rdb); err != nil {
		log.Fatal(err)
	}

	if err := u.subscribe("New", rdb); err != nil {
		log.Fatal(err)
	}

	if err := u.subscribe("Old", rdb); err != nil {
		log.Fatal(err)
	}

	if err := u.subscribe("OldPlusPlus", rdb); err != nil {
		log.Fatal(err)
	}

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
