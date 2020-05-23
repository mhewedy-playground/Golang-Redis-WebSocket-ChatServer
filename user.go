package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
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
