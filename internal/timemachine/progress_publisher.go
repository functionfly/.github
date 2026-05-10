package timemachine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisProgressPublisher struct {
	redis *redis.Client
}

func NewRedisProgressPublisher(redis *redis.Client) *RedisProgressPublisher {
	return &RedisProgressPublisher{redis: redis}
}

func (p *RedisProgressPublisher) PublishProgress(ctx context.Context, replayID uuid.UUID, data map[string]interface{}) {
	if p.redis == nil {
		return
	}
	channel := fmt.Sprintf("timemachine:progress:%s", replayID.String())
	payload, _ := json.Marshal(data)
	p.redis.Publish(ctx, channel, payload)
}
