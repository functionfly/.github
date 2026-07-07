package brain

import (
	"context"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

const (
	defaultFeedbackDays     = 7
	defaultSemanticBlend   = 0.7
	defaultHalfLifeHours   = 72.0
	defaultFreqScale       = 0.2
	defaultMaxResults      = 10
)

var (
	recencyDecay   = math.Log(2) / defaultHalfLifeHours
	signalWeights  = map[string]float64{
		"click":       1.0,
		"view":        0.5,
		"search":      1.2,
		"purchase":    1.5,
		"bookmark":    1.1,
		"dismiss":     0.3,
	}
	signalWeightsMu sync.RWMutex
)

type Scorer struct {
	repo *storage.BrainRepository
}

func NewScorer(repo *storage.BrainRepository) *Scorer {
	return &Scorer{repo: repo}
}

type ScoredSignal struct {
	Signal *storage.BrainSignal
	Score  float64
}

func (s *Scorer) ScoreSignals(_ context.Context, signals []*storage.BrainSignal, queryTime time.Time) []*ScoredSignal {
	if len(signals) == 0 {
		return nil
	}

	signalWeightsMu.RLock()
	weights := signalWeights
	signalWeightsMu.RUnlock()

	freq := make(map[string]int)
	for _, sig := range signals {
		freq[sig.EntityID]++
	}

	scored := make([]*ScoredSignal, len(signals))
	for i, sig := range signals {
		recency := recencyWeight(sig.CreatedAt, queryTime)
		frequency := frequencyWeight(freq[sig.EntityID])
		typeWeight := weights[sig.SignalType]
		if typeWeight == 0 {
			typeWeight = 1.0
		}
		score := float64(sig.Importance) * recency * frequency * typeWeight
		scored[i] = &ScoredSignal{
			Signal: sig,
			Score:  score,
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	return scored
}

func (s *Scorer) ScoreWithEmbeddings(ctx context.Context, tenantID uuid.UUID, signals []*storage.BrainSignal, queryEmbedding []float32, queryTime time.Time, maxResults int) []*ScoredSignal {
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}

	scored := s.ScoreSignals(ctx, signals, queryTime)

	if len(queryEmbedding) > 0 {
		semResults, err := s.repo.SemanticSearch(ctx, tenantID, queryEmbedding, maxResults*2)
		if err != nil {
			log.Printf("semantic search failed: %v", err)
		} else {
			semScores := make(map[string]float64)
			for _, r := range semResults {
				semScores[r.Signal.ID.String()] = r.Score
			}
			for _, ss := range scored {
				if semScore, ok := semScores[ss.Signal.ID.String()]; ok {
					ss.Score = ss.Score*defaultSemanticBlend + semScore*(1-defaultSemanticBlend)
				}
			}
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > maxResults {
		scored = scored[:maxResults]
	}

	return scored
}

func (s *Scorer) RetrainFromFeedback(ctx context.Context, repo *storage.AnalyticsEventRepository) error {
	positives, err := repo.GetFeedbackEvents(ctx, true, defaultFeedbackDays)
	if err != nil {
		return err
	}
	negatives, err := repo.GetFeedbackEvents(ctx, false, defaultFeedbackDays)
	if err != nil {
		return err
	}

	if len(positives)+len(negatives) < 10 {
		return nil
	}

	typeScores := make(map[string]float64)
	for _, e := range positives {
		typeScores[e.SignalType] += 1.0
	}
	for _, e := range negatives {
		typeScores[e.SignalType] -= 1.0
	}

	signalWeightsMu.Lock()
	defer signalWeightsMu.Unlock()

	for signalType, delta := range typeScores {
		current := signalWeights[signalType]
		if current == 0 {
			current = 1.0
		}
		adjustment := delta / float64(len(positives)+len(negatives))
		newWeight := current * (1 + adjustment)
		newWeight = math.Max(0.1, math.Min(3.0, newWeight))
		signalWeights[signalType] = newWeight
		log.Printf("updated signal weight: type=%s weight=%.3f", signalType, newWeight)
	}

	return nil
}

func recencyWeight(createdAt, queryTime time.Time) float64 {
	hours := queryTime.Sub(createdAt).Hours()
	if hours < 0 {
		hours = 0
	}
	return math.Exp(-recencyDecay * hours)
}

func frequencyWeight(count int) float64 {
	if count <= 1 {
		return 1.0
	}
	return 1.0 + math.Log2(float64(count))*defaultFreqScale
}
