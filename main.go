package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"os"
)

type user struct {
	name     string
	channels []string

	stopListener    chan bool
	listenerRunning bool

	channelsHandler *redis.PubSub
}

func newUser(name string) *user {
	return &user{
		name:         name,
		stopListener: make(chan bool),
	}
}

func (u *user) connect(rdb *redis.Client) error {
	// get all user channels (from DB) and start subscribe
	r, err := rdb.SMembers(fmt.Sprintf("user:%s:channels", u.name)).Result()
	if err != nil {
		return err
	}
	if len(r) == 0 {
		return nil
	}

	// if use has saved channels on server, then subscribe on each channel
	for _, channel := range r {
		return u.subscribe(channel, rdb)
	}

	return nil
}

func (u *user) subscribe(channel string, rdb *redis.Client) error {
	// check if already subscribed
	for i := range u.channels {
		if u.channels[i] == channel {
			return nil
		}
	}

	// save user channel to server
	userChannelsKey := fmt.Sprintf("user:%s:channels", u.name)
	if err := rdb.SAdd(userChannelsKey, channel).Err(); err != nil {
		return err
	}

	// get all user channels from server, set it as user.channels and start subscribing
	r, err := rdb.SMembers(userChannelsKey).Result()
	if err != nil {
		return err
	}
	u.channels = r

	if u.listenerRunning {
		u.stopListener <- true
	}

	u.doSubscribe(channel, rdb)

	return nil
}

func (u *user) doSubscribe(channel string, rdb *redis.Client) {
	// subscribe all channels in one request
	pubSub := rdb.Subscribe(u.channels...)
	// keep channel handler to be used in unsubscribe
	u.channelsHandler = pubSub

	// The Listener
	go func() {
		u.listenerRunning = true
		fmt.Println("starting the listener for user:", u.name, "on channels:", u.channels)
		for {
			select {
			case msg, ok := <-pubSub.Channel():
				if !ok {
					break
				}
				fmt.Println("msg:", msg.Payload, "channel:", msg.Channel)

			case <-u.stopListener:
				fmt.Println("Stop listening for user:", u.name, "on old channels")

				if err := u.channelsHandler.Unsubscribe(); err != nil {
					fmt.Fprintln(os.Stderr, "unable to unsubscribe", err)
				}
				if err := u.channelsHandler.Close(); err != nil {
					fmt.Fprintln(os.Stderr, "unable to close conn", err)
				}

				break
			}
		}
	}()
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
		1. save 2 or 3 channels under "default_channels" list on server
		on on each connect, let the user subscribe to such "default_channels"

		2. allow user to create a new channel on demand (he can invite other users by channel name)
		(make it channel:user:<user supplied name>) might make it visible with "default_channels" as well

		3. allow direct chat by create a channel with name "channel:direct:user1:user2" make both users subscribe to the channel

		4. create a list channel api to list all public channels

	*/

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
