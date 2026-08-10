package usecase

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"

	auditModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/audit/model"
)

type appVertical string

const (
	verticalGame     appVertical = "game"
	verticalCommerce appVertical = "commerce"
	verticalGeneric  appVertical = "generic"
)

var categoryLabels = map[appVertical]map[string]string{
	verticalGame: {
		"bug":     "Oyun hatası / performans",
		"feature": "Oyun içi özellik / monetization",
		"praise":  "Memnuniyet",
		"spam":    "Spam / sahte yorum",
		"other":   "Oyun deneyimi / denge",
	},
	verticalCommerce: {
		"bug":     "Uygulama / hata",
		"feature": "Ürün / özellik",
		"praise":  "Memnuniyet",
		"spam":    "Spam / gürültü",
		"other":   "Operasyon / deneyim",
	},
	verticalGeneric: {
		"bug":     "Uygulama hatası / performans",
		"feature": "Özellik talebi",
		"praise":  "Memnuniyet",
		"spam":    "Spam / gürültü",
		"other":   "Genel geri bildirim",
	},
}

var healthAppKeywords = []string{
	"kalori", "calorie", "sağlık", "health", "diyet", "diet", "fitness", "egzersiz",
	"weight", "kilo", "beslenme", "nutrition", "wellness", "macro", "yemek", "adım",
	"step", "tracker", "medical", "doctor", "hospital", "pill", "ilaç",
}

var gameKeywords = []string{
	"oyun", "game", "level", "seviye", "paywall", "reklam", "ads", "coin", "gem", "pvp",
	"multiplayer", "boss", "stage", "forge", "craft", "hile", "cheat", "gacha", "skin",
	"battle", "dungeon", "grind", "energy", "stamina", "loot", "character", "karakter",
}
var commerceKeywords = []string{
	"kargo", "iade", "beden", "sipariş", "teslimat", "ürün", "sepet", "moda", "giyim",
	"alışveriş", "shopping", "stok", "defolu", "kargo", "fatura", "kupon", "indirim",
}

func detectAppVertical(appName, clientName string, samples []auditModel.Review) appVertical {
	nameBlob := strings.ToLower(appName + " " + clientName)
	for _, kw := range healthAppKeywords {
		if strings.Contains(nameBlob, kw) {
			return verticalGeneric
		}
	}

	gameNameScore, commerceNameScore := 0, 0
	for _, kw := range gameKeywords {
		if strings.Contains(nameBlob, kw) {
			gameNameScore += 5
		}
	}
	for _, kw := range commerceKeywords {
		if strings.Contains(nameBlob, kw) {
			commerceNameScore += 5
		}
	}
	// Oyun etiketleri yalnızca uygulama adı/markası oyun olduğunda — yorumdaki "Brawl Stars" vb. tetiklemesin.
	if gameNameScore >= 5 && gameNameScore > commerceNameScore {
		return verticalGame
	}
	if commerceNameScore >= 5 && commerceNameScore > gameNameScore {
		return verticalCommerce
	}
	_ = samples
	return verticalGeneric
}

func verticalLabel(v appVertical) string {
	switch v {
	case verticalGame:
		return "mobil oyun"
	case verticalCommerce:
		return "e-ticaret / perakende uygulaması"
	default:
		return "mobil uygulama"
	}
}

func themeLabelFor(vertical appVertical, key string) string {
	if m, ok := categoryLabels[vertical]; ok {
		if label, ok := m[key]; ok {
			return label
		}
	}
	return key
}

func enrichStatistics(stats auditModel.Statistics) auditModel.Statistics {
	total := stats.TotalReviews
	if total <= 0 {
		return stats
	}

	stats.RatingDistribution = ratingDistribution(stats.RatingCounts, total)
	stats.SentimentBreakdown = sentimentFromRatings(stats.RatingCounts, total)
	stats.ThemeIntensity = themeFromCategories(stats.CategoryCounts, stats.TotalReviews, verticalGeneric)
	return stats
}

func enrichStatisticsForVertical(stats auditModel.Statistics, vertical appVertical) auditModel.Statistics {
	stats = enrichStatistics(stats)
	stats.ThemeIntensity = themeFromCategories(stats.CategoryCounts, stats.TotalReviews, vertical)
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

func themeFromCategories(categories map[string]int, totalReviews int, vertical appVertical) []auditModel.ThemeIntensity {
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
		label := themeLabelFor(vertical, it.key)
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

func fallbackRootCauses(stats auditModel.Statistics, samples []auditModel.Review, vertical appVertical) []auditModel.RootCause {
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

func fallbackReportMeta(stats auditModel.Statistics, root []auditModel.RootCause, appName string, vertical appVertical) auditModel.ReportMeta {
	current := stats.AvgRating
	goal := stretchGoalRating(current)
	gap := math.Round((goal-current)*100) / 100
	newFiveStars := estimateNewFiveStars(current, goal, stats.TotalReviews)

	topTheme := topRootTheme(root, vertical)

	meta := auditModel.ReportMeta{
		Callout: fmt.Sprintf(
			"%s (%s) için yazılı yorum ortalaması %.2f. Anlamlı iyileşme hedefi ~%.1f (fark +%.2f). "+
				"Önce tekrarlayan 1★ nedenlerini kesin; ardından memnun kullanıcıdan kontrollü 5★ toplayın.",
			appName, verticalLabel(vertical), current, goal, gap,
		),
		CurrentAvgRating:  current,
		StretchGoalRating: goal,
		Scenarios:         fallbackScenarios(current, goal, stats, newFiveStars, vertical),
		Timeline:          fallbackTimeline(topTheme, appName, vertical),
		Priorities:        fallbackPriorities(topTheme, vertical),
		ManagementFindings: fallbackFindings(stats, current, topTheme, vertical),
	}
	return meta
}

func topRootTheme(root []auditModel.RootCause, vertical appVertical) string {
	if len(root) > 0 && root[0].Theme != "" {
		return root[0].Theme
	}
	switch vertical {
	case verticalGame:
		return "Oyun dengesi / monetization"
	case verticalCommerce:
		return "Operasyon / teslimat"
	default:
		return "Uygulama kalitesi"
	}
}

func fallbackScenarios(current, goal float64, stats auditModel.Statistics, newFiveStars int, vertical appVertical) []auditModel.ImprovementScenario {
	_ = current
	_ = goal
	labelC := "Ürün fix + destek + kontrollü review"
	if vertical == verticalGame {
		labelC = "Stabilite + monetization dengesi + pozitif oturum sonrası review"
	}
	return []auditModel.ImprovementScenario{
		{
			ID: "A", Label: "Yavaş", Pace: "slow", Title: "A · Sadece yeni 5★",
			Summary: "Organik hacimle yavaş; prompt olmadan zor.", Timeline: "6–12+ ay",
			Highlight: fmt.Sprintf("~%d yeni 5★", newFiveStars),
		},
		{
			ID: "B", Label: "Orta", Pace: "mid", Title: "B · Puan güncelleme odaklı",
			Summary:  "Çözülen şikâyetlerde puan güncelleme daveti.",
			Timeline: "2–4 ay",
			Highlight: fmt.Sprintf("~%d adet düşük puan güncellemesi", minInt(stats.RatingCounts["1"], 80)),
		},
		{
			ID: "C", Label: "Önerilen", Pace: "fast", Title: "C · Karışık yol",
			Summary:  labelC,
			Timeline: "3–6 ay",
			Highlight: fmt.Sprintf("~%d güncelleme + ~%d yeni 5★", minInt(stats.RatingCounts["1"]/2, 120), maxInt(newFiveStars/4, 50)),
		},
	}
}

func fallbackTimeline(topTheme, appName string, vertical appVertical) []auditModel.TimelineItem {
	switch vertical {
	case verticalGame:
		return []auditModel.TimelineItem{
			{Horizon: "0–30 gün", Tag: "fast", Title: "Crash / performans hotfix", Body: "Crash, donma, giriş ve kayıt kaybı şikâyetlerini önceliklendirin."},
			{Horizon: "0–30 gün", Tag: "fast", Title: "Olumsuz yorum yanıtı", Body: "1–2★ yorumlara 48 saat içinde yanıt; çözüm sonrası puan güncelleme daveti."},
			{Horizon: "30–60 gün", Tag: "mid", Title: fmt.Sprintf("%s denge revizyonu", topTheme), Body: "Paywall, reklam sıklığı, progression ve ödül ekonomisini yorumlara göre ayarlayın."},
			{Horizon: "30–60 gün", Tag: "mid", Title: "Live ops / içerik", Body: "Tekrarlayan şikâyet alan etkinlik veya seviye tasarımını güncelleyin."},
			{Horizon: "60–90 gün", Tag: "mid", Title: "Pozitif oturum sonrası review", Body: "Boss/level başarısı veya uzun oturum sonrası mağaza değerlendirme prompt'u."},
			{Horizon: "Sürekli", Tag: "hard", Title: "Metrik panosu", Body: fmt.Sprintf("%s için 1★ temaları, retention ve review dönüşümünü haftalık izleyin.", appName)},
		}
	case verticalCommerce:
		return []auditModel.TimelineItem{
			{Horizon: "0–30 gün", Tag: "fast", Title: "Kurtarma ve yanıt", Body: "1–2★ yorumlara 48 saat içinde kişisel yanıt ve puan güncelleme daveti."},
			{Horizon: "0–30 gün", Tag: "fast", Title: fmt.Sprintf("%s hotfix", topTheme), Body: "Ödeme, giriş, sepet ve stok senkronu akışlarını stabilize edin."},
			{Horizon: "30–60 gün", Tag: "mid", Title: "Operasyon SLA", Body: "Kargo, iade ve ürün kalitesi için net SLA + proaktif bilgilendirme."},
			{Horizon: "30–60 gün", Tag: "mid", Title: "In-app review", Body: "Başarılı teslimat / memnuniyet sonrası mağaza değerlendirme prompt'u."},
			{Horizon: "60–90 gün", Tag: "mid", Title: "Hacim motoru", Body: "Haftalık yeni 5★ ve güncellenen puan metriklerini takip edin."},
			{Horizon: "Sürekli", Tag: "hard", Title: "Çok kanallı denge", Body: fmt.Sprintf("%s için Play ve App Store trendlerini ayrı izleyin.", appName)},
		}
	default:
		return []auditModel.TimelineItem{
			{Horizon: "0–30 gün", Tag: "fast", Title: "Stabilite ve yanıt", Body: "Crash, giriş, performans ve 1–2★ yorum yanıt süreçlerini düzeltin."},
			{Horizon: "30–60 gün", Tag: "mid", Title: fmt.Sprintf("%s iyileştirmesi", topTheme), Body: "En sık tema için ürün backlog'unu netleştirin."},
			{Horizon: "60–90 gün", Tag: "mid", Title: "Review büyümesi", Body: "Memnuniyet anında kontrollü mağaza değerlendirme prompt'u."},
		}
	}
}

func fallbackPriorities(topTheme string, vertical appVertical) []auditModel.PriorityItem {
	switch vertical {
	case verticalGame:
		return []auditModel.PriorityItem{
			{Rank: 1, Title: "Crash ve progression blocker'ları kapatın", Body: fmt.Sprintf("Öncelik: %s. Oyun deneyimi kırılmadan review toplamayın.", topTheme)},
			{Rank: 2, Title: "Monetization / paywall şikâyetlerini dengeleyin", Body: "Agresif reklam veya paywall yeni 1★ üretmeye devam eder."},
			{Rank: 3, Title: "Olumsuz yorum kurtarma", Body: "Çözüm sonrası puan güncelleme daveti hızlı skor kazandırır."},
			{Rank: 4, Title: "Pozitif oturum sonrası review", Body: "Boss/level başarısı gibi memnuniyet anında prompt kullanın."},
		}
	case verticalCommerce:
		return []auditModel.PriorityItem{
			{Rank: 1, Title: "Tekrarlayan 1★ üretimini durdurun", Body: fmt.Sprintf("Öncelik: %s.", topTheme)},
			{Rank: 2, Title: "Olumsuz yorum yanıt oranını yükseltin", Body: "Çözüm + puan güncelleme daveti hızlı skor kazandırır."},
			{Rank: 3, Title: "Teslimat sonrası review prompt", Body: "Memnun müşteriden kontrollü 5★ toplayın."},
			{Rank: 4, Title: "Haftalık skor panosu", Body: "Yeni 1★, yanıt süresi, güncellenen puan metriklerini izleyin."},
		}
	default:
		return []auditModel.PriorityItem{
			{Rank: 1, Title: "Tekrarlayan düşük puan nedenlerini kesin", Body: fmt.Sprintf("Öncelik: %s.", topTheme)},
			{Rank: 2, Title: "Olumsuz yorum yanıtı", Body: "48 saat içinde yanıt + çözüm sonrası puan güncelleme."},
			{Rank: 3, Title: "Memnuniyet anında review", Body: "Hata anında değil, başarı anında prompt."},
		}
	}
}

func fallbackFindings(stats auditModel.Statistics, current float64, topTheme string, vertical appVertical) []auditModel.ManagementFinding {
	return []auditModel.ManagementFinding{
		{Title: "Analiz kapsamı", Body: fmt.Sprintf("%d yazılı yorum; ortalama %.2f (%s).", stats.TotalReviews, current, verticalLabel(vertical))},
		{Title: "Düşük puan payı", Body: fmt.Sprintf("1–2★ yazılı payı %.0f%%.", stats.LowStarPct)},
		{Title: "En kritik tema", Body: topTheme},
	}
}

func buildFallbackSummary(a auditModel.Audit, stats auditModel.Statistics, vertical appVertical) string {
	focus := "uygulama kalitesi"
	if vertical == verticalGame {
		focus = "oyun dengesi, performans ve monetization"
	} else if vertical == verticalCommerce {
		focus = "operasyon ve ürün deneyimi"
	}
	return fmt.Sprintf(
		"%s (%s) için %d yazılı yorum analiz edildi. Ortalama puan %.2f. Öncelikli odak: %s.",
		a.AppDisplayName, verticalLabel(vertical), stats.TotalReviews, stats.AvgRating, focus,
	)
}

var foreignBrandTokens = []string{"avva", "ticimax", "lcw", "lc waikiki", "defacto"}

func mentionsForeignBrand(text, appName, clientName string) bool {
	lower := strings.ToLower(text)
	allowed := strings.ToLower(appName + " " + clientName)
	for _, brand := range foreignBrandTokens {
		if strings.Contains(lower, brand) && !strings.Contains(allowed, brand) {
			return true
		}
	}
	return false
}

func sanitizeReportMeta(meta auditModel.ReportMeta, vertical appVertical, appName string) auditModel.ReportMeta {
	if mentionsForeignBrand(meta.Callout, appName, "") {
		meta.Callout = ""
	}
	meta.Timeline = filterTimeline(meta.Timeline, vertical)
	meta.Priorities = filterPriorities(meta.Priorities, vertical)
	meta.Scenarios = filterScenarios(meta.Scenarios, vertical)
	return meta
}

func sanitizeRootCauses(in []auditModel.RootCause, vertical appVertical) []auditModel.RootCause {
	out := make([]auditModel.RootCause, 0, len(in))
	for _, rc := range in {
		if vertical == verticalGame && containsCommerceTerms(rc.Theme+" "+rc.Description) {
			continue
		}
		if vertical == verticalCommerce && containsGameTerms(rc.Theme+" "+rc.Description) && !containsCommerceTerms(rc.Theme+" "+rc.Description) {
			continue
		}
		out = append(out, rc)
	}
	return out
}

func sanitizeActionPlan(in []auditModel.ActionItem, vertical appVertical) []auditModel.ActionItem {
	out := make([]auditModel.ActionItem, 0, len(in))
	for _, item := range in {
		text := item.Action + " " + item.Title
		if vertical == verticalGame && containsCommerceTerms(text) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func containsCommerceTerms(text string) bool {
	lower := strings.ToLower(text)
	for _, kw := range commerceKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func containsGameTerms(text string) bool {
	lower := strings.ToLower(text)
	for _, kw := range gameKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func filterTimeline(items []auditModel.TimelineItem, vertical appVertical) []auditModel.TimelineItem {
	if len(items) == 0 {
		return items
	}
	out := make([]auditModel.TimelineItem, 0, len(items))
	for _, it := range items {
		text := it.Title + " " + it.Body
		if vertical == verticalGame && containsCommerceTerms(text) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func filterPriorities(items []auditModel.PriorityItem, vertical appVertical) []auditModel.PriorityItem {
	out := make([]auditModel.PriorityItem, 0, len(items))
	for _, it := range items {
		if vertical == verticalGame && containsCommerceTerms(it.Title+" "+it.Body) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func filterScenarios(items []auditModel.ImprovementScenario, vertical appVertical) []auditModel.ImprovementScenario {
	out := make([]auditModel.ImprovementScenario, 0, len(items))
	for _, it := range items {
		if vertical == verticalGame && containsCommerceTerms(it.Title+" "+it.Summary) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func normalizeToken(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
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

func parseFeedbackFromRaw(raw string) ([]auditModel.CategoryInsight, []auditModel.FeedbackSuggestion, []auditModel.FeedbackSuggestion, []auditModel.FeaturedReview) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil, nil, nil, nil
	}
	var payload struct {
		CategoryInsights   []auditModel.CategoryInsight   `json:"category_insights"`
		FeatureSuggestions []auditModel.FeedbackSuggestion `json:"feature_suggestions"`
		BugSuggestions     []auditModel.FeedbackSuggestion `json:"bug_suggestions"`
		FeaturedReviews    []auditModel.FeaturedReview    `json:"featured_reviews"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &payload); err != nil {
		return nil, nil, nil, nil
	}
	return payload.CategoryInsights, payload.FeatureSuggestions, payload.BugSuggestions, payload.FeaturedReviews
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
