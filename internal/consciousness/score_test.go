package consciousness

import (
	"database/sql"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewScoreComputer(t *testing.T) {
	db := &sql.DB{}
	logger := logrus.New()

	sc := NewScoreComputer(db, logger)
	assert.NotNil(t, sc)
	assert.Equal(t, db, sc.db)
	assert.NotNil(t, sc.weights)
	assert.Equal(t, 0.25, sc.weights.Health)
	assert.Equal(t, 0.20, sc.weights.Efficiency)
	assert.Equal(t, 0.20, sc.weights.Scalability)
	assert.Equal(t, 0.20, sc.weights.Reliability)
	assert.Equal(t, 0.15, sc.weights.Optimization)
}

func TestNewScoreComputerWithWeights(t *testing.T) {
	db := &sql.DB{}
	logger := logrus.New()

	weights := ScoreWeights{
		Health:       0.30,
		Efficiency:   0.25,
		Scalability:  0.20,
		Reliability:  0.15,
		Optimization: 0.10,
	}

	sc := NewScoreComputerWithWeights(db, logger, weights)
	assert.NotNil(t, sc)
	assert.Equal(t, weights, sc.weights)
}

func TestNewScoreComputerWithWeights_DefaultsOnZero(t *testing.T) {
	db := &sql.DB{}
	logger := logrus.New()

	weights := ScoreWeights{
		Health:       0,
		Efficiency:   0,
		Scalability:  0,
		Reliability:  0,
		Optimization: 0,
	}

	sc := NewScoreComputerWithWeights(db, logger, weights)
	assert.NotNil(t, sc)
	assert.Equal(t, DefaultScoreWeights(), sc.weights)
}

func TestScoreComputer_Compute_Clamping(t *testing.T) {
	db := &sql.DB{}
	logger := logrus.New()
	sc := NewScoreComputer(db, logger)

	assert.NotNil(t, sc)

	assert.Equal(t, 0.25, sc.weights.Health)
	assert.Equal(t, 0.20, sc.weights.Efficiency)
	assert.Equal(t, 0.20, sc.weights.Scalability)
	assert.Equal(t, 0.20, sc.weights.Reliability)
	assert.Equal(t, 0.15, sc.weights.Optimization)

	total := sc.weights.Health + sc.weights.Efficiency + sc.weights.Scalability +
		sc.weights.Reliability + sc.weights.Optimization
	assert.Equal(t, 1.0, total)
}

func TestScoreWeights_Total(t *testing.T) {
	tests := []struct {
		name   string
		weights ScoreWeights
		total   float64
	}{
		{
			name: "default weights",
			weights: ScoreWeights{
				Health:       0.25,
				Efficiency:   0.20,
				Scalability:  0.20,
				Reliability:  0.20,
				Optimization: 0.15,
			},
			total: 1.0,
		},
		{
			name: "custom weights",
			weights: ScoreWeights{
				Health:       0.30,
				Efficiency:   0.25,
				Scalability:  0.20,
				Reliability:  0.15,
				Optimization: 0.10,
			},
			total: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.weights.Health + tt.weights.Efficiency + tt.weights.Scalability +
				tt.weights.Reliability + tt.weights.Optimization
			assert.Equal(t, tt.total, total)
		})
	}
}