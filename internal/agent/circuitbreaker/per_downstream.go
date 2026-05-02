package circuitbreaker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type DownstreamConfig struct {
	DownstreamID         string
	FailureThreshold    int
	SuccessThreshold    int
	CooldownDuration    time.Duration
	HalfOpenMaxRequests int
}

type PerDownstreamBreaker struct {
	mu            sync.RWMutex
	breakers      map[string]*Breaker
	configs       map[string]Config
	redis         *redis.Client
	agentID       string
	onStateChange func(agentID, downstreamID string, from, to State)
}

func NewPerDownstreamBreaker(agentID string, redisClient *redis.Client) *PerDownstreamBreaker {
	return &PerDownstreamBreaker{
		agentID:  agentID,
		breakers: make(map[string]*Breaker),
		configs:  make(map[string]Config),
		redis:    redisClient,
	}
}

func (p *PerDownstreamBreaker) RegisterDownstream(ctx context.Context, downstreamID string, config Config) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 2
	}
	if config.CooldownDuration == 0 {
		config.CooldownDuration = 30 * time.Second
	}
	if config.HalfOpenMaxRequests == 0 {
		config.HalfOpenMaxRequests = 1
	}

	stateChangeHandler := func(from, to State) {
		if p.onStateChange != nil {
			p.onStateChange(p.agentID, downstreamID, from, to)
		}
	}
	config.OnStateChange = stateChangeHandler

	p.configs[downstreamID] = config
	p.breakers[downstreamID] = New(config)

	if p.redis != nil {
		return p.persistConfig(ctx, downstreamID, config)
	}
	return nil
}

type persistableConfig struct {
	FailureThreshold    int           `json:"failure_threshold"`
	SuccessThreshold    int           `json:"success_threshold"`
	CooldownDuration    time.Duration `json:"cooldown_duration"`
	HalfOpenMaxRequests int           `json:"half_open_max_requests"`
}

func (p *PerDownstreamBreaker) persistConfig(ctx context.Context, downstreamID string, config Config) error {
	key := fmt.Sprintf("agent:%s:circuit_breaker:%s:config", p.agentID, downstreamID)
	persistCfg := persistableConfig{
		FailureThreshold:    config.FailureThreshold,
		SuccessThreshold:    config.SuccessThreshold,
		CooldownDuration:   config.CooldownDuration,
		HalfOpenMaxRequests: config.HalfOpenMaxRequests,
	}
	data, err := json.Marshal(persistCfg)
	if err != nil {
		return err
	}
	return p.redis.Set(ctx, key, data, 24*time.Hour).Err()
}

func (p *PerDownstreamBreaker) LoadConfig(ctx context.Context, downstreamID string) (Config, error) {
	if p.redis == nil {
		return DefaultConfig(), nil
	}

	key := fmt.Sprintf("agent:%s:circuit_breaker:%s:config", p.agentID, downstreamID)
	data, err := p.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}

	var persistCfg persistableConfig
	if err := json.Unmarshal(data, &persistCfg); err != nil {
		return DefaultConfig(), err
	}
	return Config{
		FailureThreshold:    persistCfg.FailureThreshold,
		SuccessThreshold:    persistCfg.SuccessThreshold,
		CooldownDuration:   persistCfg.CooldownDuration,
		HalfOpenMaxRequests: persistCfg.HalfOpenMaxRequests,
	}, nil
}

func (p *PerDownstreamBreaker) GetBreaker(downstreamID string) *Breaker {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.breakers[downstreamID]
}

func (p *PerDownstreamBreaker) Execute(ctx context.Context, downstreamID string, fn func() error) error {
	p.mu.RLock()
	breaker, exists := p.breakers[downstreamID]
	p.mu.RUnlock()

	if !exists {
		p.mu.Lock()
		breaker, exists = p.breakers[downstreamID]
		if !exists {
			config := Config{
				FailureThreshold:    5,
				SuccessThreshold:    2,
				CooldownDuration:    30 * time.Second,
				HalfOpenMaxRequests: 1,
			}
			stateChangeHandler := func(from, to State) {
				if p.onStateChange != nil {
					p.onStateChange(p.agentID, downstreamID, from, to)
				}
			}
			config.OnStateChange = stateChangeHandler
			breaker = New(config)
			p.breakers[downstreamID] = breaker
			if p.redis != nil {
				_ = p.persistConfig(ctx, downstreamID, config)
			}
		}
		p.mu.Unlock()
	}

	return breaker.Execute(fn)
}

func (p *PerDownstreamBreaker) Allow(downstreamID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if breaker, exists := p.breakers[downstreamID]; exists {
		return breaker.Allow()
	}
	return true
}

func (p *PerDownstreamBreaker) Record(downstreamID string, err error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if breaker, exists := p.breakers[downstreamID]; exists {
		breaker.Record(err)
	}
}

func (p *PerDownstreamBreaker) GetState(downstreamID string) State {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if breaker, exists := p.breakers[downstreamID]; exists {
		return breaker.State()
	}
	return StateClosed
}

func (p *PerDownstreamBreaker) GetAllStates() map[string]State {
	p.mu.RLock()
	defer p.mu.RUnlock()
	states := make(map[string]State)
	for id, breaker := range p.breakers {
		states[id] = breaker.State()
	}
	return states
}

func (p *PerDownstreamBreaker) Reset(downstreamID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if breaker, exists := p.breakers[downstreamID]; exists {
		breaker.Reset()
	}
}

func (p *PerDownstreamBreaker) ResetAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, breaker := range p.breakers {
		breaker.Reset()
	}
}

func (p *PerDownstreamBreaker) RemoveDownstream(downstreamID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.breakers, downstreamID)
	delete(p.configs, downstreamID)
}

func (p *PerDownstreamBreaker) SetStateChangeHandler(handler func(agentID, downstreamID string, from, to State)) {
	p.onStateChange = handler
}

func (p *PerDownstreamBreaker) ListDownstreams() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	downstreams := make([]string, 0, len(p.breakers))
	for id := range p.breakers {
		downstreams = append(downstreams, id)
	}
	return downstreams
}