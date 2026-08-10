package usecase

import (
	"fmt"
	"sort"
	"strings"

	auditModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/audit/model"
)

var categoryOrder = []string{"bug", "feature", "other", "praise", "spam"}

func toQuote(rv auditModel.Review) auditModel.ReviewQuote {
	return auditModel.ReviewQuote{
		Text:      truncate(rv.Text, 220),
		Rating:    rv.Rating,
		Store:     rv.Store,
		Sentiment: rv.Sentiment,
	}
}

func buildCategoryInsights(stats auditModel.Statistics, reviews []auditModel.Review, vertical appVertical) []auditModel.CategoryInsight {
	total := stats.TotalReviews
	if total <= 0 {
		total = len(reviews)
	}
	byCat := map[string][]auditModel.Review{}
	for _, rv := range reviews {
		cat := rv.Category
		if cat == "" {
			cat = "other"
		}
		byCat[cat] = append(byCat[cat], rv)
	}

	out := make([]auditModel.CategoryInsight, 0, len(categoryOrder))
	for _, cat := range categoryOrder {
		count := stats.CategoryCounts[cat]
		if count == 0 {
			count = len(byCat[cat])
		}
		if count == 0 {
			continue
		}
		pct := 0.0
		if total > 0 {
			pct = roundPct(float64(count) / float64(total) * 100)
		}
		quotes := sampleQuotes(byCat[cat], 3)
		out = append(out, auditModel.CategoryInsight{
			Category: cat,
			Label:    themeLabelFor(vertical, cat),
			Count:    count,
			Pct:      pct,
			Reviews:  quotes,
		})
	}
	return out
}

func buildFeatureSuggestions(reviews []auditModel.Review) []auditModel.FeedbackSuggestion {
	var featureReviews []auditModel.Review
	for _, rv := range reviews {
		if rv.Category == "feature" {
			featureReviews = append(featureReviews, rv)
		}
	}
	sort.Slice(featureReviews, func(i, j int) bool {
		return len(featureReviews[i].Text) > len(featureReviews[j].Text)
	})
	return buildSuggestionsFromReviews(featureReviews, "feature", 5)
}

func buildBugSuggestions(reviews []auditModel.Review) []auditModel.FeedbackSuggestion {
	var bugReviews []auditModel.Review
	for _, rv := range reviews {
		if rv.Category == "bug" || (rv.Rating <= 2 && rv.Category != "praise" && rv.Category != "spam") {
			bugReviews = append(bugReviews, rv)
		}
	}
	sort.Slice(bugReviews, func(i, j int) bool {
		if bugReviews[i].Rating != bugReviews[j].Rating {
			return bugReviews[i].Rating < bugReviews[j].Rating
		}
		return len(bugReviews[i].Text) > len(bugReviews[j].Text)
	})
	return buildSuggestionsFromReviews(bugReviews, "bug", 5)
}

func buildSuggestionsFromReviews(reviews []auditModel.Review, kind string, max int) []auditModel.FeedbackSuggestion {
	out := make([]auditModel.FeedbackSuggestion, 0, max)
	used := map[string]bool{}
	for _, rv := range reviews {
		if len(out) >= max {
			break
		}
		title := suggestionTitle(rv.Text, kind)
		key := strings.ToLower(title)
		if used[key] {
			continue
		}
		used[key] = true
		priority := "P1"
		if kind == "bug" && rv.Rating <= 2 {
			priority = "P0"
		}
		out = append(out, auditModel.FeedbackSuggestion{
			Title:             title,
			Summary:           truncate(rv.Text, 160),
			Category:          kind,
			Priority:          priority,
			SupportingReviews: []auditModel.ReviewQuote{toQuote(rv)},
		})
	}
	return out
}

func buildFeaturedReviews(reviews []auditModel.Review) []auditModel.FeaturedReview {
	type scored struct {
		rv    auditModel.Review
		score int
	}
	items := make([]scored, 0, len(reviews))
	for _, rv := range reviews {
		if strings.TrimSpace(rv.Text) == "" {
			continue
		}
		score := len(rv.Text) / 10
		if rv.Rating <= 2 {
			score += 20
		} else if rv.Rating >= 4 {
			score += 12
		}
		if rv.Category == "bug" || rv.Category == "feature" {
			score += 8
		}
		items = append(items, scored{rv: rv, score: score})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].score > items[j].score })

	out := make([]auditModel.FeaturedReview, 0, 6)
	neg, pos := 0, 0
	for _, it := range items {
		if len(out) >= 6 {
			break
		}
		rv := it.rv
		if rv.Rating <= 2 && neg >= 3 {
			continue
		}
		if rv.Rating >= 4 && pos >= 3 {
			continue
		}
		highlight := "Kullanıcı geri bildirimi"
		if rv.Rating <= 2 {
			highlight = "Dikkat çeken şikâyet"
			neg++
		} else if rv.Rating >= 4 {
			highlight = "Olumlu öne çıkan yorum"
			pos++
		}
		out = append(out, auditModel.FeaturedReview{
			Text:      truncate(rv.Text, 260),
			Rating:    rv.Rating,
			Store:     rv.Store,
			Category:  rv.Category,
			Sentiment: rv.Sentiment,
			Highlight: highlight,
		})
	}
	return out
}

func computeStoreStats(reviews []auditModel.Review) auditModel.Statistics {
	stats := auditModel.Statistics{
		CategoryCounts:  map[string]int{},
		SentimentCounts: map[string]int{},
		RatingCounts:    map[string]int{},
	}
	var ratingSum int
	for _, rv := range reviews {
		stats.TotalReviews++
		key := fmt.Sprintf("%d", rv.Rating)
		stats.RatingCounts[key]++
		ratingSum += rv.Rating
		cat := rv.Category
		if cat == "" {
			cat = "other"
		}
		stats.CategoryCounts[cat]++
		if rv.Sentiment != "" {
			stats.SentimentCounts[rv.Sentiment]++
		}
	}
	if stats.TotalReviews > 0 {
		stats.AvgRating = float64(ratingSum) / float64(stats.TotalReviews)
		low := stats.RatingCounts["1"] + stats.RatingCounts["2"]
		stats.LowStarPct = roundPct(float64(low) / float64(stats.TotalReviews) * 100)
	}
	return enrichStatistics(stats)
}

func buildStoreBreakdown(a auditModel.Audit, classified []auditModel.Review, vertical appVertical) []auditModel.StoreInsight {
	type storeSpec struct {
		key   string
		label string
	}
	var specs []storeSpec
	if strings.TrimSpace(a.PlayAppID) != "" {
		specs = append(specs, storeSpec{"play", "Google Play"})
	}
	if strings.TrimSpace(a.AppStoreAppID) != "" {
		specs = append(specs, storeSpec{"appstore", "App Store"})
	}
	if len(specs) <= 1 {
		return nil
	}

	out := make([]auditModel.StoreInsight, 0, len(specs))
	for _, spec := range specs {
		subset := make([]auditModel.Review, 0)
		for _, rv := range classified {
			if rv.Store == spec.key {
				subset = append(subset, rv)
			}
		}
		if len(subset) == 0 {
			continue
		}
		storeStats := computeStoreStats(subset)
		if spec.key == "play" {
			storeStats.PlayCount = len(subset)
		} else {
			storeStats.AppStoreCount = len(subset)
		}
		storeStats = enrichStatisticsForVertical(storeStats, vertical)
		out = append(out, auditModel.StoreInsight{
			Store:              spec.key,
			Label:              spec.label,
			Statistics:         storeStats,
			CategoryInsights:   buildCategoryInsights(storeStats, subset, vertical),
			FeatureSuggestions: buildFeatureSuggestions(subset),
			BugSuggestions:     buildBugSuggestions(subset),
			FeaturedReviews:    buildFeaturedReviews(subset),
		})
	}
	return out
}

func sampleQuotes(reviews []auditModel.Review, n int) []auditModel.ReviewQuote {
	if len(reviews) == 0 {
		return nil
	}
	sort.Slice(reviews, func(i, j int) bool { return len(reviews[i].Text) > len(reviews[j].Text) })
	out := make([]auditModel.ReviewQuote, 0, n)
	for _, rv := range reviews {
		if len(out) >= n {
			break
		}
		out = append(out, toQuote(rv))
	}
	return out
}

func suggestionTitle(text, kind string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		if kind == "feature" {
			return "Yeni özellik talebi"
		}
		return "Hata / performans şikâyeti"
	}
	first := text
	if idx := strings.IndexAny(first, ".!?\n"); idx > 20 && idx < 100 {
		first = first[:idx]
	}
	first = truncate(first, 90)
	if kind == "feature" {
		return fmt.Sprintf("%s (feature)", first)
	}
	return fmt.Sprintf("%s (bug)", first)
}

func mergeFeedbackSuggestions(base, fromLLM []auditModel.FeedbackSuggestion, kind string) []auditModel.FeedbackSuggestion {
	if len(fromLLM) == 0 {
		return base
	}
	out := make([]auditModel.FeedbackSuggestion, 0, len(fromLLM))
	for _, s := range fromLLM {
		if s.Category == "" {
			s.Category = kind
		}
		if strings.TrimSpace(s.Title) == "" {
			continue
		}
		if len(s.SupportingReviews) == 0 && strings.TrimSpace(s.Summary) != "" {
			s.SupportingReviews = []auditModel.ReviewQuote{{Text: truncate(s.Summary, 220)}}
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return base
	}
	return out
}

func mergeFeaturedReviews(base, fromLLM []auditModel.FeaturedReview) []auditModel.FeaturedReview {
	if len(fromLLM) == 0 {
		return base
	}
	return fromLLM
}

func applyVerticalCategoryLabels(insights []auditModel.CategoryInsight, vertical appVertical) []auditModel.CategoryInsight {
	for i := range insights {
		cat := insights[i].Category
		if cat == "" {
			cat = "other"
		}
		insights[i].Label = themeLabelFor(vertical, cat)
	}
	return insights
}

func mergeCategoryInsights(base, fromLLM []auditModel.CategoryInsight) []auditModel.CategoryInsight {
	if len(fromLLM) == 0 {
		return base
	}
	byCat := map[string]auditModel.CategoryInsight{}
	for _, c := range base {
		byCat[c.Category] = c
	}
	for _, c := range fromLLM {
		if existing, ok := byCat[c.Category]; ok {
			if len(c.Reviews) == 0 {
				c.Reviews = existing.Reviews
			}
			if c.Count == 0 {
				c.Count = existing.Count
			}
			if c.Pct == 0 {
				c.Pct = existing.Pct
			}
			c.Label = existing.Label
		}
		byCat[c.Category] = c
	}
	out := make([]auditModel.CategoryInsight, 0, len(categoryOrder))
	for _, cat := range categoryOrder {
		if c, ok := byCat[cat]; ok {
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return base
	}
	return out
}
