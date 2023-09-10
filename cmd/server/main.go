package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/bells307/go-tsrq/internal/queue"
	tsrq_grpc "github.com/bells307/go-tsrq/internal/transport/grpc"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	globalCtx := context.Background()

	q := queue.NewRedisTimestampRoundQueue(client, "gotsr", time.Second*10, time.Second*10)
	q.RunCleaner(globalCtx, 5*time.Second)

	grpcServer := grpc.NewServer()
	handler := tsrq_grpc.NewTSRQHandler(q)
	handler.Register(grpcServer)

	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal(err)
	}

	reflection.Register(grpcServer)

	log.Println("start grpc server")
	err = grpcServer.Serve(ln)
	if err != nil {
		log.Fatalf("can't create grpc listener: %v", err)
	}
}
