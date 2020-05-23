package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"os"
)

const userChannelFmt = "user:%s:channels"

type user struct {
	name            string
	channels        []string
	channelsHandler *redis.PubSub

	stopListener    chan bool
	listenerRunning bool
}

func newUser(name string) *user {
	return &user{
		name:         name,
		stopListener: make(chan bool),
	}
}

func (u *user) subscribe(rdb *redis.Client, channel string) error {

	userChannelsKey := fmt.Sprintf(userChannelFmt, u.name)

	if rdb.SIsMember(userChannelsKey, channel).Val() {
		return nil
	}
	if err := rdb.SAdd(userChannelsKey, channel).Err(); err != nil {
		return err
	}

	return u.connect(rdb)
}

func (u *user) unsubscribe(rdb *redis.Client, channel string) error {

	userChannelsKey := fmt.Sprintf(userChannelFmt, u.name)

	if !rdb.SIsMember(userChannelsKey, channel).Val() {
		return nil
	}
	if err := rdb.SRem(userChannelsKey, channel).Err(); err != nil {
		return err
	}

	return u.connect(rdb)
}

func (u *user) connect(rdb *redis.Client) error {
	// get all user channels (from DB) and start subscribe
	c, err := rdb.SMembers(fmt.Sprintf("user:%s:channels", u.name)).Result()
	if err != nil {
		return err
	}
	if len(c) == 0 {
		fmt.Println("no channels to connect to for user: ", u.name)
		return nil
	}

	if u.channelsHandler != nil {
		if err := u.channelsHandler.Unsubscribe(); err != nil {
			fmt.Fprintln(os.Stderr, "unable to unsubscribe", err)
		}
		if err := u.channelsHandler.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "unable to close conn", err)
		}
	}
	if u.listenerRunning {
		u.stopListener <- true
	}

	u.channels = c

	return u.doConnect(rdb)
}

func (u *user) doConnect(rdb *redis.Client) error {
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
				break
			}
		}
	}()

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
		1. save 2 or 3 channels under "default_channels" list on server
		on on each connect, let the user subscribe to such "default_channels"

		2. allow user to create a new channel on demand (he can invite other users by channel name)
		(make it channel:user:<user supplied name>) might make it visible with "default_channels" as well

		3. allow direct chat by create a channel with name "channel:direct:user1:user2" make both users subscribe to the channel

		4. create a list channel api to list all public channels

	*/

	//if err := u.subscribe(rdb, "ch1"); err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := u.subscribe(rdb, "ch2"); err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := u.subscribe(rdb, "ch3"); err != nil {
	//	log.Fatal(err)
	//}
	//
	//if err := u.subscribe(rdb, "ch4"); err != nil {
	//	log.Fatal(err)
	//}

	if err := u.unsubscribe(rdb, "ch4"); err != nil {
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
