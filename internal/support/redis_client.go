package support

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisClientAdapter adapts a go-redis client to the minimal RedisClient
// interface used by the support service.
type RedisClientAdapter struct {
	client *redis.Client
}

func NewRedisClientAdapter(client *redis.Client) *RedisClientAdapter {
	return &RedisClientAdapter{client: client}
}

func (a *RedisClientAdapter) Publish(ctx context.Context, channel string, message interface{}) error {
	return a.client.Publish(ctx, channel, message).Err()
}

func (a *RedisClientAdapter) Subscribe(ctx context.Context, channel string) (<-chan string, error) {
	pubsub := a.client.Subscribe(ctx, channel)

	out := make(chan string)
	ch := pubsub.Channel()

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logrus.WithFields(logrus.Fields{
					"panic":  rec,
					"stack":  fmt.Sprintf("%v", rec),
					"channel": channel,
				}).Error("Redis subscription goroutine panicked")
			}
		}()
		defer close(out)
		defer func() {
			_ = pubsub.Close()
		}()

		for msg := range ch {
			out <- msg.Payload
		}
	}()

	return out, nil
}

