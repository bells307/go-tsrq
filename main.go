package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bells307/go-tsrq/internal/queue"
	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()

	q := queue.NewRedisTimestampRoundQueue(client, "gotsr", time.Second*10, time.Second*10)

	err := q.Enqueue(ctx, "123", "somedata")
	if err != nil {
		panic(err)
	}

	err = q.Enqueue(ctx, "456", "somedata2")
	if err != nil {
		panic(err)
	}

	err = q.Enqueue(ctx, "789", "somedata3")
	if err != nil {
		panic(err)
	}

	q.RunCleaner(ctx, time.Second*10)

	time.Sleep(time.Second * 60)

	res, err := q.Dequeue(ctx)

	fmt.Println(res)

	if err != nil {
		panic(err)
	}
}
