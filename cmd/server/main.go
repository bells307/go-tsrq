package main

import (
	"context"
	"log"
	"net"

	"github.com/bells307/go-tsrq/cmd/server/config"
	"github.com/bells307/go-tsrq/internal/queue"
	tsrq_grpc "github.com/bells307/go-tsrq/internal/transport/grpc"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("error while loading application configuration: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	globalCtx := context.Background()

	log.Println("initializing queue")
	q := queue.NewRedisTimestampRoundQueue(client, cfg.Queue.Name, cfg.Queue.DeqPeriod, cfg.Queue.Ttl)
	q.RunCleaner(globalCtx, cfg.Queue.CleanPeriod)

	grpcServer := grpc.NewServer()
	handler := tsrq_grpc.NewTSRQHandler(q)
	handler.Register(grpcServer)

	ln, err := net.Listen("tcp", cfg.GrpcListen)
	if err != nil {
		log.Fatal(err)
	}

	reflection.Register(grpcServer)

	log.Println("start grpc server on", cfg.GrpcListen)

	err = grpcServer.Serve(ln)
	if err != nil {
		log.Fatalf("can't create grpc listener: %v", err)
	}
}
