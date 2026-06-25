package companyranking

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const cacheTTL = 5 * time.Minute

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) key(prefix, country, category string) string {
	return fmt.Sprintf("companyrank:%s:%s:%s", prefix, country, category)
}

func (c *Cache) GetLeaderboard(ctx context.Context, country, category string) ([]Ranking, bool, error) {
	if c.rdb == nil {
		return nil, false, nil
	}
	data, err := c.rdb.Get(ctx, c.key("lb", country, category)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var rows []Ranking
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, false, err
	}
	return rows, true, nil
}

func (c *Cache) SetLeaderboard(ctx context.Context, country, category string, rows []Ranking) error {
	if c.rdb == nil {
		return nil
	}
	data, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, c.key("lb", country, category), data, cacheTTL).Err()
}
