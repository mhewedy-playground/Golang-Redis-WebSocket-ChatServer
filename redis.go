package main

import (
	"github.com/gomodule/redigo/redis"
	"log"
	"os"
)

var pool *redis.Pool

func initPool() {
	pool = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				log.Printf("Error: failed init redis pool: %s", err.Error())
				os.Exit(1)
			}
			return conn, err
		},
	}
}
