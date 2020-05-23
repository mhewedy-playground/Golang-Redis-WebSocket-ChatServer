package main

import (
	"github.com/go-redis/redis/v7"
	"log"
)

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
