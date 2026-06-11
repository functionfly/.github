package consciousness

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadRedundancyThresholds(t *testing.T) {
	os.Setenv(RedundancySimilarityThresholdEnv, "0.90")
	defer os.Unsetenv(RedundancySimilarityThresholdEnv)

	threshold := loadRedundancySimilarityThreshold()
	assert.Equal(t, 0.90, threshold)
}

func TestLoadRedundancyCoOccurrenceMin(t *testing.T) {
	os.Setenv(RedundancyCoOccurrenceMinEnv, "100")
	defer os.Unsetenv(RedundancyCoOccurrenceMinEnv)

	min := loadRedundancyCoOccurrenceMin()
	assert.Equal(t, 100, min)
}

func TestLoadRedundancyMaxPairs(t *testing.T) {
	os.Setenv(RedundancyMaxPairsEnv, "10")
	defer os.Unsetenv(RedundancyMaxPairsEnv)

	max := loadRedundancyMaxPairs()
	assert.Equal(t, 10, max)
}

func TestLoadRedundancyThresholds_Invalid(t *testing.T) {
	os.Setenv(RedundancySimilarityThresholdEnv, "invalid")
	defer os.Unsetenv(RedundancySimilarityThresholdEnv)

	threshold := loadRedundancySimilarityThreshold()
	assert.Equal(t, DefaultRedundancySimilarityThreshold, threshold)
}

func TestLoadRedundancyThresholds_OutOfRange(t *testing.T) {
	os.Setenv(RedundancySimilarityThresholdEnv, "1.5")
	defer os.Unsetenv(RedundancySimilarityThresholdEnv)

	threshold := loadRedundancySimilarityThreshold()
	assert.Equal(t, DefaultRedundancySimilarityThreshold, threshold)

	os.Setenv(RedundancySimilarityThresholdEnv, "-0.1")
	defer os.Unsetenv(RedundancySimilarityThresholdEnv)

	threshold = loadRedundancySimilarityThreshold()
	assert.Equal(t, DefaultRedundancySimilarityThreshold, threshold)
}

func TestDefaultRedundancyConstants(t *testing.T) {
	assert.Equal(t, 0.82, DefaultRedundancySimilarityThreshold)
	assert.Equal(t, 50, DefaultRedundancyCoOccurrenceMin)
	assert.Equal(t, 5, DefaultRedundancyMaxPairs)
}

func TestLoadMarketplaceSimilarityThreshold(t *testing.T) {
	os.Setenv(MarketplaceSimilarityThresholdEnv, "0.85")
	defer os.Unsetenv(MarketplaceSimilarityThresholdEnv)

	threshold := loadMarketplaceSimilarityThreshold()
	assert.Equal(t, 0.85, threshold)
}

func TestLoadMarketplaceMaxInsights(t *testing.T) {
	os.Setenv(MarketplaceMaxInsightsEnv, "10")
	defer os.Unsetenv(MarketplaceMaxInsightsEnv)

	max := loadMarketplaceMaxInsights()
	assert.Equal(t, 10, max)
}

func TestLoadMarketplaceThresholds_Invalid(t *testing.T) {
	os.Setenv(MarketplaceSimilarityThresholdEnv, "invalid")
	defer os.Unsetenv(MarketplaceSimilarityThresholdEnv)

	threshold := loadMarketplaceSimilarityThreshold()
	assert.Equal(t, DefaultMarketplaceSimilarityThreshold, threshold)
}

func TestDefaultMarketplaceConstants(t *testing.T) {
	assert.Equal(t, 0.75, DefaultMarketplaceSimilarityThreshold)
	assert.Equal(t, 5, DefaultMarketplaceMaxInsights)
}

func TestRedundancyAnalyzer_Configurable(t *testing.T) {
	os.Setenv(RedundancySimilarityThresholdEnv, "0.95")
	os.Setenv(RedundancyMaxPairsEnv, "3")
	defer func() {
		os.Unsetenv(RedundancySimilarityThresholdEnv)
		os.Unsetenv(RedundancyMaxPairsEnv)
	}()

	analyzer := NewRedundancyAnalyzer(nil, nil)
	assert.Equal(t, 0.95, analyzer.similarityThreshold)
	assert.Equal(t, 3, analyzer.maxPairs)
}

func TestMarketplaceAnalyzer_Configurable(t *testing.T) {
	os.Setenv(MarketplaceSimilarityThresholdEnv, "0.90")
	os.Setenv(MarketplaceMaxInsightsEnv, "8")
	defer func() {
		os.Unsetenv(MarketplaceSimilarityThresholdEnv)
		os.Unsetenv(MarketplaceMaxInsightsEnv)
	}()

	analyzer := NewMarketplaceAnalyzer(nil, nil)
	assert.Equal(t, 0.90, analyzer.similarityThreshold)
	assert.Equal(t, 8, analyzer.maxInsights)
}