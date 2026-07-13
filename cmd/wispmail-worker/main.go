package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/wispmail/wispmail/config"
	"github.com/wispmail/wispmail/internal/queue"
	"github.com/wispmail/wispmail/internal/storage/postgres"
	"github.com/wispmail/wispmail/internal/storage/redis"
)

func main() {
	cfg := config.MustLoad("")

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	db, err := postgres.NewClient(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.NewClient(ctx, cfg.Redis.URL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	worker, err := queue.NewWorker(redisClient, db, queue.WorkerConfig{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}

	log.Println("Starting Wispmail worker...")

	if err := worker.Start(ctx); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	<-ctx.Done()
	log.Println("Shutting down worker...")

	if err := worker.Stop(context.Background()); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}