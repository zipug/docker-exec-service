package redis

import (
	"context"
	"encoding/json"
	"errors"
	"executor/internal/core/models"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrPing = errors.New("could not ping database")

type RedisRepository struct {
	rdb    *redis.Client
	config *redis.Options
}

type RepositoryConsumer struct {
	client       *RedisRepository
	subscription *redis.PubSub
}

type customHandler func(models.BotMessage) error

func NewRepositoryConsumer(host, password string, port, db int) *RepositoryConsumer {
	client := NewRedisRepository(host, password, port, db)
	return &RepositoryConsumer{client: client}
}

func NewRedisRepository(host, password string, port, db int) *RedisRepository {
	repo := &RedisRepository{}
	if err := repo.InvokeConnect(host, password, port, db); err != nil {
		e := fmt.Errorf("REDIS: redis://default:%s@%s:%d/%d\n%w", password, host, port, db, err)
		panic(e)
	}
	return repo
}

func (repo *RedisRepository) InvokeConnect(host, password string, port, db int) error {
	conf := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
		Protocol: 2,
	}
	rdb := redis.NewClient(&conf)
	repo.config = &conf
	repo.rdb = rdb
	if err := repo.PingTest(); err != nil {
		panic(err)
	}
	return nil
}

func (repo *RedisRepository) PingTest() error {
	max_errs := 5
	errs := 0
	timeout := 1 * time.Second
	for max_errs > 0 {
		if err := repo.rdb.Ping(context.Background()).Err(); err != nil {
			fmt.Printf("could not ping database: %s\n", err.Error())
			fmt.Printf("retrying in %s\n", timeout)
			max_errs--
			errs++
			time.Sleep(timeout)
		}
		max_errs = 0
		errs = 0
	}
	if errs == 0 {
		return nil
	}
	return fmt.Errorf("%w: redis_uri: %s", ErrPing, repo.config.Addr)
}

func (c *RepositoryConsumer) ConsumerMessages(ctx context.Context, queue_names []string, handler customHandler) {
	for _, queue := range queue_names {
		switch queue {
		case "bot":
			go c.handleMessage(ctx, queue, handler)
		default:
			fmt.Println("unsupported queue type")
		}
	}
}

func (c *RepositoryConsumer) handleMessage(ctx context.Context, queue string, handler customHandler) {
	consumerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	fmt.Printf("subscribing to queue: %s\n", queue)
	c.subscription = c.client.rdb.Subscribe(consumerCtx, queue)
	defer c.subscription.Close()

	channel := c.subscription.Channel()

	fmt.Printf("[%s] consumer started listening...\n", queue)
	for {
		select {
		case <-consumerCtx.Done():
			fmt.Printf("[%s] consumer stopped listening...\n", queue)
			return
		case msg := <-channel:
			var message models.BotMessage
			err := json.Unmarshal([]byte(msg.Payload), &message)
			if err != nil {
				fmt.Printf("[%s] could not unmarshal message: %s\n", queue, err.Error())
				continue
			}
			if err := handler(message); err != nil {
				fmt.Printf("[%s] %s\n", queue, err.Error())
				continue
			}
		}
	}
}
