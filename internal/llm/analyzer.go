package llm

import (
	"strings"
	"time"
)

type Result struct {
	Category  string 
	Sentiment string 
	RawOutput string 
	LatencyMs int    
}

type Analyzer interface {
	Analyze(text string) (Result, error)
}

type MockAnalyzer struct{}

func NewMockAnalyzer() *MockAnalyzer {
	return &MockAnalyzer{}
}

func (m *MockAnalyzer) Analyze(text string) (Result, error) {
	start := time.Now()
	lower := strings.ToLower(text)

	category := "other"
	switch {
	case containsAny(lower, "crash", "bug", "error", "freeze", "broken"):
		category = "bug"
	case containsAny(lower, "please add", "would be nice", "feature", "wish", "suggestion"):
		category = "feature"
	case containsAny(lower, "love", "great", "awesome", "amazing", "best"):
		category = "praise"
	case containsAny(lower, "http", "www", "buy now", "promo", "click here"):
		category = "spam"
	}

	sentiment := "neutral"
	switch {
	case containsAny(lower, "love", "great", "awesome", "amazing", "good", "best"):
		sentiment = "positive"
	case containsAny(lower, "hate", "bad", "worst", "terrible", "crash", "broken", "awful"):
		sentiment = "negative"
	}

	return Result{
		Category:  category,
		Sentiment: sentiment,
		RawOutput: "mock: category=" + category + ", sentiment=" + sentiment,
		LatencyMs: int(time.Since(start).Milliseconds()),
	}, nil
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}