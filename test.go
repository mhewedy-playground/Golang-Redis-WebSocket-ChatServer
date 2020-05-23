package main

import (
	"github.com/go-redis/redis/v7"
	"log"
)

func main() {

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	pubSub := rdb.Subscribe("A")
	pubSub = rdb.Subscribe("B")
	pubSub = rdb.Subscribe("C")
	pubSub2 := rdb.Subscribe("D")
	pubSub = rdb.Subscribe("E")
	pubSub = rdb.Subscribe("F")

	if err := pubSub.Unsubscribe(); err != nil {
		log.Fatal(err)
	}

	if err := pubSub2.Unsubscribe(); err != nil {
		log.Fatal(err)
	}

	select {}

}
