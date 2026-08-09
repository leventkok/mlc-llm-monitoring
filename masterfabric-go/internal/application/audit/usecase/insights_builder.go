package usecase

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	auditModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/audit/model"
)

var categoryLabels = map[string]string{
	"bug":     "Uygulama / hata",
	"feature": "Özellik / ürün",
	"praise":  "Memnuniyet",
	"spam":    "Spam / gürültü",
	"other":   "Diğer",
}

func enrichStatistics(stats auditModel.Statistics) auditModel.Statistics {
	total := stats.TotalReviews
	if total <= 0 {
		return stats
	}

	stats.RatingDistribution = ratingDistribution(stats.RatingCounts, total)
	stats.SentimentBreakdown = sentimentFromRatings(stats.RatingCounts, total)
	stats.ThemeIntensity = themeFromCategories(stats.CategoryCounts, stats.TotalReviews)
	return stats
}

func ratingDistribution(counts map[string]int, total int) []auditModel.RatingBucket {
	out := make([]auditModel.RatingBucket, 0, 5)
	for star := 1; star <= 5; star++ {
		key := fmt.Sprintf("%d", star)
		count := counts[key]
		out = append(out, auditModel.RatingBucket{
			Star:  star,
			Count: count,
			Pct:   roundPct(float64(count) / float64(total) * 100),
		})
	}
	return out
}

func sentimentFromRatings(counts map[string]int, total int) map[string]auditModel.SentimentBucket {
	neg := counts["1"] + counts["2"]
	neu := counts["3"]
	pos := counts["4"] + counts["5"]
	return map[string]auditModel.SentimentBucket{
		"negative": {Count: neg, Pct: roundPct(float64(neg) / float64(total) * 100)},
		"neutral":  {Count: neu, Pct: roundPct(float64(neu) / float64(total) * 100)},
		"positive": {Count: pos, Pct: roundPct(float64(pos) / float64(total) * 100)},
	}
}

func themeFromCategories(categories map[string]int, totalReviews int) []auditModel.ThemeIntensity {
	type pair struct {
		key   string
		count int
	}
	items := make([]pair, 0, len(categories))
	negTotal := 0
	for k, v := range categories {
		if k == "praise" || k == "spam" {
			continue
		}
		items = append(items, pair{key: k, count: v})
		negTotal += v
	}
	sort.Slice(items, func(i, j int) bool { return items[i].count > items[j].count })
	if negTotal == 0 {
		negTotal = totalReviews
		if negTotal == 0 {
			negTotal = 1
		}
	}
	maxCount := 1
	if len(items) > 0 {
		maxCount = items[0].count
		if maxCount == 0 {
			maxCount = 1
		}
	}
	out := make([]auditModel.ThemeIntensity, 0, len(items))
	for _, it := range items {
		label := categoryLabels[it.key]
		if label == "" && it.key != "" {
			label = strings.ToUpper(it.key[:1]) + it.key[1:]
		} else if label == "" {
			label = "Diğer"
		}
		out = append(out, auditModel.ThemeIntensity{
			Theme: it.key,
			Label: label,
			Count: it.count,
			Pct:   roundPct(float64(it.count) / float64(maxCount) * 100),
		})
	}
	return out
}

func stretchGoalRating(current float64) float64 {
	if current <= 0 {
		return 4.0
	}
	goal := current + 0.45
	if goal > 5 {
		goal = 5
	}
	return math.Round(goal*10) / 10
}

func fallbackRootCauses(stats auditModel.Statistics, samples []auditModel.Review) []auditModel.RootCause {
	if len(stats.ThemeIntensity) > 0 {
		out := make([]auditModel.RootCause, 0, len(stats.ThemeIntensity))
		for i, t := range stats.ThemeIntensity {
			if i >= 7 {
				break
			}
			out = append(out, auditModel.RootCause{
				Theme:         t.Label,
				Description:   fmt.Sprintf("%s teması %d yazılı yorumda öne çıkıyor.", t.Label, t.Count),
				AffectedRating: "1-3",
				SampleCount:   t.Count,
			})
		}
		return out
	}
	out := make([]auditModel.RootCause, 0, 3)
	for i, rv := range samples {
		if i >= 3 {
			break
		}
		out = append(out, auditModel.RootCause{
			Theme:         rv.Category,
			Description:   truncate(rv.Text, 160),
			AffectedRating: fmt.Sprintf("%d", rv.Rating),
			SampleCount:   1,
			Examples:      []string{truncate(rv.Text, 120)},
		})
	}
	return out
}

func fallbackReportMeta(stats auditModel.Statistics, root []auditModel.RootCause, appName string) auditModel.ReportMeta {
	current := stats.AvgRating
	goal := stretchGoalRating(current)
	gap := math.Round((goal-current)*100) / 100
	newFiveStars := estimateNewFiveStars(current, goal, stats.TotalReviews)

	topTheme := "ürün ve operasyon"
	if len(root) > 0 && root[0].Theme != "" {
		topTheme = root[0].Theme
	}

	return auditModel.ReportMeta{
		Callout: fmt.Sprintf(
			"Yazılı yorum ortalaması %.2f. Anlamlı iyileşme için odak hedef ~%.1f (fark +%.2f). "+
				"Sadece yeni 5★ ile bu mesafe kabaca ~%d değerlendirme gerektirir; düşük puan güncellemesi + ürün fix karışık yol daha hızlıdır.",
			current, goal, gap, newFiveStars,
		),
		CurrentAvgRating:  current,
		StretchGoalRating: goal,
		Scenarios: []auditModel.ImprovementScenario{
			{
				ID: "A", Label: "Yavaş", Pace: "slow", Title: "A · Sadece yeni 5★",
				Summary:  "Organik hacimle yavaş; prompt olmadan zor.",
				Timeline: "6–12+ ay",
				Highlight: fmt.Sprintf("~%d yeni 5★", newFiveStars),
			},
			{
				ID: "B", Label: "Orta", Pace: "mid", Title: "B · Puan güncelleme odaklı",
				Summary:  "Mevcut düşük puanlı kullanıcıları çözüm sonrası güncellemeye davet edin.",
				Timeline: "2–4 ay",
				Highlight: fmt.Sprintf("~%d adet 1★→5★ etkisi hedeflenir", minInt(stats.RatingCounts["1"], 80)),
			},
			{
				ID: "C", Label: "Önerilen", Pace: "fast", Title: "C · Karışık yol",
				Summary:  "Ürün fix + destek kurtarma + kontrollü in-app review.",
				Timeline: "3–6 ay",
				Highlight: fmt.Sprintf("~%d güncelleme + ~%d yeni 5★", minInt(stats.RatingCounts["1"]/2, 120), newFiveStars/4),
			},
		},
		Timeline: []auditModel.TimelineItem{
			{Horizon: "0–30 gün", Tag: "fast", Title: "Kurtarma ve yanıt", Body: "1–2★ yorumlara 48 saat içinde yanıt; çözülen vakada puan güncelleme daveti."},
			{Horizon: "0–30 gün", Tag: "fast", Title: fmt.Sprintf("%s hotfix", topTheme), Body: "En sık şikâyet temalarında crash, giriş, ödeme akışlarını stabilize edin."},
			{Horizon: "30–60 gün", Tag: "mid", Title: "Operasyon SLA", Body: "Kargo, iade ve stok senkronu için net SLA + proaktif bilgilendirme."},
			{Horizon: "30–60 gün", Tag: "mid", Title: "In-app review", Body: "Başarılı teslimat / memnuniyet sonrası mağaza değerlendirme prompt'u."},
			{Horizon: "60–90 gün", Tag: "mid", Title: "Hacim motoru", Body: "Haftalık yeni 5★ ve güncellenen puan metriklerini takip edin."},
			{Horizon: "Sürekli", Tag: "hard", Title: "Çok kanallı denge", Body: fmt.Sprintf("%s için Play ve App Store trendlerini ayrı izleyin.", appName)},
		},
		Priorities: []auditModel.PriorityItem{
			{Rank: 1, Title: "Tekrarlayan 1★ üretimini durdurun", Body: fmt.Sprintf("Öncelik: %s. Fix olmadan toplanan 5★'ler yeni 1★'lerle nötrlenir.", topTheme)},
			{Rank: 2, Title: "Olumsuz yorum yanıt oranını yükseltin", Body: "Çözüm + puan güncelleme daveti, yeni 5★ toplamaktan hızlı skor kazandırır."},
			{Rank: 3, Title: "Memnun kullanıcıdan kontrollü 5★", Body: "Hata anında değil; başarı anında review prompt kullanın."},
			{Rank: 4, Title: "Haftalık skor panosu", Body: "Yeni 1★, yanıt süresi, güncellenen puan, prompt sonrası 5★ metriklerini izleyin."},
		},
		ManagementFindings: []auditModel.ManagementFinding{
			{Title: "Yazılı yorum ortalaması mağaza algısından farklı olabilir", Body: fmt.Sprintf("Analiz edilen %d yazılı yorum ortalaması %.2f; olumsuzlar metin bırakma eğilimindedir.", stats.TotalReviews, current)},
			{Title: "Düşük puan payı", Body: fmt.Sprintf("1–2★ yazılı payı %.0f%%.", stats.LowStarPct)},
			{Title: "En kritik tema", Body: topTheme},
		},
	}
}

func fallbackActionPlan(meta auditModel.ReportMeta) []auditModel.ActionItem {
	out := make([]auditModel.ActionItem, 0, len(meta.Timeline))
	for _, t := range meta.Timeline {
		out = append(out, auditModel.ActionItem{
			Priority:       "P1",
			HorizonDays:    t.Horizon,
			Action:         t.Body,
			Title:          t.Title,
			Tag:            t.Tag,
			ExpectedImpact: "Skor ve memnuniyet iyileşmesi",
			OwnerHint:      "Ürün + destek",
		})
	}
	return out
}

func estimateNewFiveStars(current, goal float64, total int) int {
	if total <= 0 || goal <= current {
		return 100
	}
	// Approximate stars needed: (goal*total - current*total) / (5 - goal) simplified
	num := (goal - current) * float64(total)
	den := 5.0 - goal
	if den <= 0.1 {
		den = 0.1
	}
	return int(math.Ceil(num / den))
}

func parseInsightBundle(raw string) (auditModel.ReportMeta, []auditModel.RootCause, []auditModel.ActionItem, string, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return auditModel.ReportMeta{}, nil, nil, "", fmt.Errorf("no json object")
	}
	var payload struct {
		ExecutiveSummary   string                      `json:"executive_summary"`
		Callout            string                      `json:"callout"`
		RootCauses         []auditModel.RootCause      `json:"root_causes"`
		Scenarios          []auditModel.ImprovementScenario `json:"scenarios"`
		Timeline           []auditModel.TimelineItem   `json:"timeline"`
		Priorities         []auditModel.PriorityItem   `json:"priorities"`
		ManagementFindings []auditModel.ManagementFinding `json:"management_findings"`
		ActionPlan         []auditModel.ActionItem     `json:"action_plan"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &payload); err != nil {
		return auditModel.ReportMeta{}, nil, nil, "", err
	}
	meta := auditModel.ReportMeta{
		Callout:            payload.Callout,
		Scenarios:          payload.Scenarios,
		Timeline:           payload.Timeline,
		Priorities:         payload.Priorities,
		ManagementFindings: payload.ManagementFindings,
	}
	return meta, payload.RootCauses, payload.ActionPlan, payload.ExecutiveSummary, nil
}

func mergeReportMeta(base, fromLLM auditModel.ReportMeta, stats auditModel.Statistics) auditModel.ReportMeta {
	base.CurrentAvgRating = stats.AvgRating
	base.StretchGoalRating = stretchGoalRating(stats.AvgRating)
	if fromLLM.Callout != "" {
		base.Callout = fromLLM.Callout
	}
	if len(fromLLM.Scenarios) > 0 {
		base.Scenarios = fromLLM.Scenarios
	}
	if len(fromLLM.Timeline) > 0 {
		base.Timeline = fromLLM.Timeline
	}
	if len(fromLLM.Priorities) > 0 {
		base.Priorities = fromLLM.Priorities
	}
	if len(fromLLM.ManagementFindings) > 0 {
		base.ManagementFindings = fromLLM.ManagementFindings
	}
	return base
}

func roundPct(v float64) float64 {
	return math.Round(v*10) / 10
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
