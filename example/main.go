package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gbenroscience/webredis/sessions"
	"github.com/go-redis/redis/v8"
)

func main() {

	client, err := redisConnInit("127.0.0.1", 6379, "whatsyourpassword?")

	if err != nil {
		panic(err)
	}

	redisSessionStore := sessions.NewWebRedisStore(client, "I believe in God! He inspires me", 7200)

	//Somewhere in your application
	var r http.Request
	// Assume this request is from your HandlerFunc and is initialized of course:

	sess, err := redisSessionStore.Get(&r, "user")

	//etc.

}

func redisConnInit(redisAddr string, redisPort int, password string) (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     redisAddr + ":" + strconv.Itoa(redisPort),
		Password: password,
		DB:       0,
	})
	xxx := client.Ping(context.Background())

	res, err := xxx.Result()
	fmt.Printf("redis: Name() = %s, FullName() = %s, err = %v, err2 = %v, result = %s val = %s\n", xxx.Name(), xxx.FullName(), xxx.Err(), err, res, xxx.Val())

	//return redisStore
	return client, err
}
