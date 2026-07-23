package usecase

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/repository"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/validate"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/metrics"
)

// ComputeAutoQuality scores LLM output quality (1-5) from raw response metadata.
// This measures output health (format, latency), not human ground-truth accuracy.
func ComputeAutoQuality(category, sentiment, rawOutput string, latencyMs int) int {
	quality := 1

	if validate.Category(category) != nil || validate.Sentiment(sentiment) != nil {
		return quality
	}

	raw := strings.TrimSpace(rawOutput)
	if raw == "" {
		return 2
	}

	quality = 3

	if parsedCategory, parsedSentiment, ok := parseClassificationJSON(raw); ok {
		quality = 4
		if parsedCategory == category && parsedSentiment == sentiment {
			quality = 5
		}
	}

	switch {
	case latencyMs <= 0:
		quality = clampQuality(quality - 1)
	case latencyMs > 10_000:
		quality = clampQuality(quality - 2)
	case latencyMs > 5_000:
		quality = clampQuality(quality - 1)
	}

	return clampQuality(quality)
}

func parseClassificationJSON(raw string) (category, sentiment string, ok bool) {
	idx := strings.Index(raw, "{")
	if idx < 0 {
		return "", "", false
	}
	end := strings.LastIndex(raw, "}")
	if end <= idx {
		return "", "", false
	}

	var obj struct {
		Category  string `json:"category"`
		Sentiment string `json:"sentiment"`
	}
	if err := json.Unmarshal([]byte(raw[idx:end+1]), &obj); err != nil {
		return "", "", false
	}
	if validate.Category(obj.Category) != nil || validate.Sentiment(obj.Sentiment) != nil {
		return "", "", false
	}
	return obj.Category, obj.Sentiment, true
}

func clampQuality(v int) int {
	if v < 1 {
		return 1
	}
	if v > 5 {
		return 5
	}
	return v
}

func persistAutoScore(ctx context.Context, reviews repository.ReviewRepository, userID string, decision model.Decision) error {
	quality := ComputeAutoQuality(decision.Category, decision.Sentiment, decision.RawOutput, decision.LatencyMs)
	metrics.RecordAutoScore(quality)
	return reviews.CreateScore(ctx, model.Score{
		ID:         uuid.NewString(),
		DecisionID: decision.ID,
		Quality:    quality,
		ScoredBy:   userID,
	}, userID)
}
