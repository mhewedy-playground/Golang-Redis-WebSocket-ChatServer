package user

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"time"
)

const (
	// used to track users that used chat. mainly for listing users in the /users api, in real world chat app
	// such user list should be separated into user management module.
	usersKey       = "users"
	userChannelFmt = "user:%s:channels"
	ChannelsKey    = "channels"
)

//rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
var rdb = redis.NewClient(&redis.Options{
	Addr:         "localhost:6379",
	DialTimeout:  10 * time.Second,
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
	PoolSize:     2,
	PoolTimeout:  30 * time.Second,
})

type User struct {
	Name            string
	channelsHandler *redis.PubSub

	stopListenerChan chan struct{}
	listening        bool

	MessageChan chan redis.Message
}

//Connect connect user to user channels on redis
func Connect(name string) (*User, error) {

	if _, err := rdb.SAdd(usersKey, name).Result(); err != nil {
		return nil, err
	}

	u := &User{
		Name:             name,
		stopListenerChan: make(chan struct{}),
		MessageChan:      make(chan redis.Message),
	}

	if err := u.connect(); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) Subscribe(channel string) error {

	userChannelsKey := fmt.Sprintf(userChannelFmt, u.Name)

	if rdb.SIsMember(userChannelsKey, channel).Val() {
		return nil
	}
	if err := rdb.SAdd(userChannelsKey, channel).Err(); err != nil {
		return err
	}

	return u.connect()
}

func (u *User) Unsubscribe(channel string) error {

	userChannelsKey := fmt.Sprintf(userChannelFmt, u.Name)

	if !rdb.SIsMember(userChannelsKey, channel).Val() {
		return nil
	}
	if err := rdb.SRem(userChannelsKey, channel).Err(); err != nil {
		return err
	}

	return u.connect()
}

func (u *User) connect() error {

	var c []string

	log.Println("User Connect Invoke.")

	c1, err := rdb.SMembers(ChannelsKey).Result()
	if err != nil {
		return err
	}
	c = append(c, c1...)

	// get all user channels (from DB) and start subscribe
	c2, err := rdb.SMembers(fmt.Sprintf(userChannelFmt, u.Name)).Result()
	if err != nil {
		return err
	}
	c = append(c, c2...)

	if len(c) == 0 {
		log.Println("no channels to connect to for user: ", u.Name)
		return nil
	}

	if u.channelsHandler != nil {
		if err := u.channelsHandler.Unsubscribe(); err != nil {
			return err
		}
		if err := u.channelsHandler.Close(); err != nil {
			return err
		}
	}
	if u.listening {
		u.stopListenerChan <- struct{}{}
	}

	return u.doConnect(c...)
}

func (u *User) doConnect(channels ...string) error {
	// subscribe all channels in one request
	pubSub := rdb.Subscribe(channels...)
	// keep channel handler to be used in unsubscribe
	u.channelsHandler = pubSub

	// The Listener
	go func() {
		u.listening = true
		log.Println("starting the listener for user:", u.Name, "on channels:", channels)

		if err := Chat("general", u.Name+" Connected"); err != nil {
			log.Printf("User Connect Error: %s \n", err)
		}

		for {
			select {
			case msg, ok := <-pubSub.Channel():
				if !ok {
					return
				}
				u.MessageChan <- *msg

			case <-u.stopListenerChan:
				log.Println("stopping the listener for user:", u.Name)
				return
			}
		}
	}()
	return nil
}

func (u *User) Disconnect(userName string) error {
	if u.channelsHandler != nil {
		if err := u.channelsHandler.Unsubscribe(); err != nil {
			return err
		}
		if err := u.channelsHandler.Close(); err != nil {
			return err
		}
	}
	if u.listening {
		u.stopListenerChan <- struct{}{}
	}

	if err := rdb.SRem(usersKey, userName).Err(); err != nil {
		return err
	}

	close(u.MessageChan)

	if err := Chat("general", u.Name+" DisConnected"); err != nil {
		log.Printf("User Connect Error: %s \n", err)
	}

	return nil
}

func Chat(channel string, content string) error {
	return rdb.Publish(channel, content).Err()
}

func List() ([]string, error) {
	return rdb.SMembers(usersKey).Result()
}

func GetChannels(username string) ([]string, error) {

	if !rdb.SIsMember(usersKey, username).Val() {
		return nil, errors.New("user not exists")
	}

	var c []string

	c1, err := rdb.SMembers(ChannelsKey).Result()
	if err != nil {
		return nil, err
	}
	c = append(c, c1...)

	// get all user channels (from DB) and start subscribe
	c2, err := rdb.SMembers(fmt.Sprintf(userChannelFmt, username)).Result()
	if err != nil {
		return nil, err
	}
	c = append(c, c2...)

	return c, nil
}
