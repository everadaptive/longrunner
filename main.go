package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

var client *redis.Client

func main() {
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		fmt.Print(err)
	}

	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				client.Publish("mychannel2", "hello").Result()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/callback", callback)
	http.ListenAndServe(":8080", nil)
}

func callback(w http.ResponseWriter, req *http.Request) {
	fmt.Println("listening to mychannel2")

	pubsub := client.Subscribe("mychannel2")
	defer pubsub.Close()

	for i := 0; i < 10; i++ {
		// ReceiveTimeout is a low level API. Use ReceiveMessage instead.
		msgi, err := pubsub.ReceiveTimeout(30 * time.Second)
		if err != nil {
			fmt.Print(err)
			break
		}

		switch msg := msgi.(type) {
		case *redis.Subscription:
			fmt.Println("subscribed to", msg.Channel)
		case *redis.Message:
			fmt.Println("received", msg.Payload, "from", msg.Channel)
			fmt.Fprintf(w, "%s", msg.Payload)
			return
		default:
			panic("unreached")
		}
	}
}

func startHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(10 * time.Second)
	http.Redirect(w, req, "http://localhost:8080/callback", http.StatusMovedPermanently)
}
