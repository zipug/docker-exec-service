package main

import (
	"context"
	"executor/internal/core/config"
	"executor/internal/docker"
	"executor/internal/repository/postgres"
	"executor/internal/repository/redis"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var queue_names = []string{"bot"}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	defer stop()
	cfg := config.NewConfigService()
	repo := postgres.NewPostgresRepository(cfg)
	docker := docker.NewDockerService(ctx, repo, cfg)
	consumer := redis.NewRepositoryConsumer(
		cfg.Redis.Host,
		cfg.Redis.RedisPassword,
		cfg.Redis.Port,
		cfg.Redis.DB,
	)
	go func() {
		consumer.ConsumerMessages(ctx, queue_names, docker.DockerFactory)
	}()
	select {
	case <-ctx.Done():
		if err := docker.StopAllContainers(context.Background()); err != nil {
			fmt.Println(err)
		}
		fmt.Println("shutting down...")
	}
}
