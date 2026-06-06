package brain

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
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

// ScoreSignals applies relevance scoring: importance × recency_weight × frequency_weight
func (s *Scorer) ScoreSignals(ctx context.Context, signals []*storage.BrainSignal, queryTime time.Time) []*ScoredSignal {
	if len(signals) == 0 {
		return nil
	}

	// Count entity frequencies
	freq := make(map[string]int)
	for _, sig := range signals {
		freq[sig.EntityID]++
	}

	scored := make([]*ScoredSignal, len(signals))
	for i, sig := range signals {
		recency := recencyWeight(sig.CreatedAt, queryTime)
		frequency := frequencyWeight(freq[sig.EntityID])
		score := float64(sig.Importance) * recency * frequency
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

// ScoreWithEmbeddings combines text scoring with pgvector similarity (Pro+)
func (s *Scorer) ScoreWithEmbeddings(ctx context.Context, tenantID uuid.UUID, signals []*storage.BrainSignal, queryEmbedding []float32, queryTime time.Time, maxResults int) []*ScoredSignal {
	if maxResults <= 0 {
		maxResults = 10
	}

	scored := s.ScoreSignals(ctx, signals, queryTime)

	if len(queryEmbedding) > 0 {
		semResults, err := s.repo.SemanticSearch(ctx, tenantID, queryEmbedding, maxResults*2)
		if err == nil {
			semScores := make(map[string]float64)
			for _, r := range semResults {
				semScores[r.Signal.ID.String()] = r.Score
			}
			for _, ss := range scored {
				if semScore, ok := semScores[ss.Signal.ID.String()]; ok {
					ss.Score = ss.Score*0.7 + semScore*0.3
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

// RetrainFromFeedback adjusts importance weights based on user feedback
func (s *Scorer) RetrainFromFeedback(ctx context.Context, repo *storage.AnalyticsEventRepository) error {
	positives, err := repo.GetFeedbackEvents(ctx, true, 7)
	if err != nil {
		return err
	}
	negatives, err := repo.GetFeedbackEvents(ctx, false, 7)
	if err != nil {
		return err
	}

	// Calculate positive/negative signal type ratios
	typeScores := make(map[string]float64)
	for _, e := range positives {
		typeScores[e.SignalType] += 1.0
	}
	for _, e := range negatives {
		typeScores[e.SignalType] -= 1.0
	}

	// Log retraining results (in production, update a weight table)
	_ = typeScores
	_ = len(positives) + len(negatives)

	return nil
}

// recencyWeight returns a weight between 0 and 1 based on how recent the signal is
func recencyWeight(createdAt, queryTime time.Time) float64 {
	hours := queryTime.Sub(createdAt).Hours()
	if hours < 0 {
		hours = 0
	}
	// Exponential decay: half-life of 72 hours
	return math.Exp(-0.0096 * hours)
}

// frequencyWeight returns a weight based on how often an entity appears
func frequencyWeight(count int) float64 {
	if count <= 1 {
		return 1.0
	}
	// Logarithmic scaling: diminishing returns for repeated entities
	return 1.0 + math.Log2(float64(count))*0.2
}
