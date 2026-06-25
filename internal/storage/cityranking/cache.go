package cityranking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheTTL is the read-side cache lifetime for leaderboard responses. The
// hourly recompute job warms this cache as it finishes.
const CacheTTL = 5 * time.Minute

// Cache wraps a Redis client with city-ranking-specific helpers. The cache is
// best-effort — all errors are returned to the caller so the handler can fall
// back to a direct DB read.
type Cache struct {
	rdb *redis.Client
}

// NewCache creates a Cache from a redis client. A nil client is allowed and
// produces a no-op cache (every call returns a cache miss).
func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func leaderboardKey(category Category) string {
	return fmt.Sprintf("cityrank:leaderboard:%s", category)
}
func leaderboardCountryKey(country string, category Category) string {
	return fmt.Sprintf("cityrank:leaderboard:country:%s:%s", country, category)
}
func metroKey(slug string, category Category) string {
	return fmt.Sprintf("cityrank:metro:%s:%s", slug, category)
}
func moversKey(direction string, category Category) string {
	return fmt.Sprintf("cityrank:movers:%s:%s", direction, category)
}
func myCityKey(userID string) string { return fmt.Sprintf("cityrank:me:%s", userID) }
func statesKey(country string, category Category) string {
	if country == "" {
		return fmt.Sprintf("cityrank:states:%s", category)
	}
	return fmt.Sprintf("cityrank:states:country:%s:%s", country, category)
}
func stateKey(country, code string, category Category) string {
	return fmt.Sprintf("cityrank:state:%s:%s:%s", strings.ToUpper(country), strings.ToUpper(code), category)
}
func mapKey(category Category) string { return fmt.Sprintf("cityrank:map:%s", category) }

// GetLeaderboard returns the cached top-100 leaderboard for a category.
func (c *Cache) GetLeaderboard(ctx context.Context, country string, category Category) ([]Ranking, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, false, nil
	}
	key := leaderboardKey(category)
	if country != "" {
		key = leaderboardCountryKey(country, category)
	}
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out []Ranking
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// SetLeaderboard stores the top-100 leaderboard for a category.
func (c *Cache) SetLeaderboard(ctx context.Context, country string, category Category, rankings []Ranking) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(rankings)
	if err != nil {
		return err
	}
	key := leaderboardKey(category)
	if country != "" {
		key = leaderboardCountryKey(country, category)
	}
	return c.rdb.Set(ctx, key, raw, CacheTTL).Err()
}

// GetMetro returns a cached single-metro response for a category.
func (c *Cache) GetMetro(ctx context.Context, slug string, category Category) (*Ranking, []Ranking, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, nil, false, nil
	}
	raw, err := c.rdb.Get(ctx, metroKey(slug, category)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	var entry struct {
		Current *Ranking  `json:"current"`
		History []Ranking `json:"history"`
	}
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, nil, false, err
	}
	return entry.Current, entry.History, true, nil
}

// SetMetro caches a single-metro response for a category.
func (c *Cache) SetMetro(ctx context.Context, slug string, category Category, current *Ranking, history []Ranking) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(struct {
		Current *Ranking  `json:"current"`
		History []Ranking `json:"history"`
	}{current, history})
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, metroKey(slug, category), raw, CacheTTL).Err()
}

// GetMovers returns a cached movers list for a category.
func (c *Cache) GetMovers(ctx context.Context, direction string, category Category) ([]Ranking, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, false, nil
	}
	raw, err := c.rdb.Get(ctx, moversKey(direction, category)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out []Ranking
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// SetMovers caches a movers list for a category.
func (c *Cache) SetMovers(ctx context.Context, direction string, category Category, rankings []Ranking) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(rankings)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, moversKey(direction, category), raw, CacheTTL).Err()
}

// GetMyCity returns a cached "my city" response for a user.
func (c *Cache) GetMyCity(ctx context.Context, userID string) (*Ranking, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, false, nil
	}
	raw, err := c.rdb.Get(ctx, myCityKey(userID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out Ranking
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

// SetMyCity caches a "my city" response.
func (c *Cache) SetMyCity(ctx context.Context, userID string, ranking *Ranking) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(ranking)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, myCityKey(userID), raw, CacheTTL).Err()
}

// InvalidateAll wipes every cityrank:* key. Called by the recompute job
// after a successful cycle.
func (c *Cache) InvalidateAll(ctx context.Context) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	iter := c.rdb.Scan(ctx, 0, "cityrank:*", 200).Iterator()
	keys := []string{}
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 200 {
			if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.rdb.Del(ctx, keys...).Err()
	}
	return nil
}

// ── State rankings ────────────────────────────────────────────────────────

// GetStates returns the cached state leaderboard for a country (or global
// when country is empty) and category.
func (c *Cache) GetStates(ctx context.Context, country string, category Category) ([]StateRanking, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, false, nil
	}
	raw, err := c.rdb.Get(ctx, statesKey(country, category)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out []StateRanking
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// SetStates caches a state leaderboard for a category.
func (c *Cache) SetStates(ctx context.Context, country string, category Category, rows []StateRanking) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, statesKey(country, category), raw, CacheTTL).Err()
}

// GetState returns a cached single-state aggregate for a category.
func (c *Cache) GetState(ctx context.Context, country, code string, category Category) (*StateRanking, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, false, nil
	}
	raw, err := c.rdb.Get(ctx, stateKey(country, code, category)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var s StateRanking
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, false, err
	}
	return &s, true, nil
}

// SetState caches a single-state aggregate for a category.
func (c *Cache) SetState(ctx context.Context, country, code string, category Category, s *StateRanking) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, stateKey(country, code, category), raw, CacheTTL).Err()
}

// ── Map points ────────────────────────────────────────────────────────────

// GetMapPoints returns the cached map points for a category.
func (c *Cache) GetMapPoints(ctx context.Context, category Category) ([]MapPoint, bool, error) {
	if c == nil || c.rdb == nil {
		return nil, false, nil
	}
	raw, err := c.rdb.Get(ctx, mapKey(category)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var out []MapPoint
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

// SetMapPoints caches the map points for a category.
func (c *Cache) SetMapPoints(ctx context.Context, category Category, pts []MapPoint) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(pts)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, mapKey(category), raw, CacheTTL).Err()
}
