package main

import (
	"github.com/go-redis/redis/v7"
	"net/http"
)

func h(rdb *redis.Client, fn func(http.ResponseWriter, *http.Request, *redis.Client)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, rdb)
	}
}

func connectHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {

}

func subscribeHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {

}

func unsubscribeHandler(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {

}
