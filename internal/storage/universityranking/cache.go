package universityranking

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheTTL is the read-side cache lifetime for leaderboard responses. The
// hourly recompute job warms this cache as it finishes.
const CacheTTL = 5 * time.Minute

// Cache wraps a Redis client with university-ranking-specific helpers. The
// cache is best-effort — all errors are returned to the caller so the
// handler can fall back to a direct DB read.
type Cache struct {
	rdb *redis.Client
}

// NewCache creates a Cache from a redis client. A nil client is allowed
// and produces a no-op cache (every call returns a cache miss).
func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func leaderboardKey(country string, category Category) string {
	if country == "" {
		country = "GLOBAL"
	}
	return fmt.Sprintf("unirank:leaderboard:%s:%s", country, category)
}
func detailKey(slug string, category Category) string {
	return fmt.Sprintf("unirank:detail:%s:%s", slug, category)
}

func (c *Cache) GetLeaderboard(ctx context.Context, country string, category Category) ([]Ranking, bool, error) {
	if c.rdb == nil {
		return nil, false, nil
	}
	raw, err := c.rdb.Get(ctx, leaderboardKey(country, category)).Result()
	if err != nil || raw == "" {
		return nil, false, err
	}
	var out []Ranking
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (c *Cache) SetLeaderboard(ctx context.Context, country string, category Category, rows []Ranking) error {
	if c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, leaderboardKey(country, category), raw, CacheTTL).Err()
}

func (c *Cache) GetDetail(ctx context.Context, slug string, category Category) (*Ranking, bool, error) {
	if c.rdb == nil {
		return nil, false, nil
	}
	raw, err := c.rdb.Get(ctx, detailKey(slug, category)).Result()
	if err != nil || raw == "" {
		return nil, false, err
	}
	var rk Ranking
	if err := json.Unmarshal([]byte(raw), &rk); err != nil {
		return nil, false, err
	}
	return &rk, true, nil
}

func (c *Cache) SetDetail(ctx context.Context, slug string, category Category, rk *Ranking) error {
	if c.rdb == nil || rk == nil {
		return nil
	}
	raw, err := json.Marshal(rk)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, detailKey(slug, category), raw, CacheTTL).Err()
}
