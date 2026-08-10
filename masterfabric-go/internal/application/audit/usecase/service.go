package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	auditDTO "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/audit/dto"
	llmScope "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/scope"
	auditModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/audit/model"
	pgAudit "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/postgres/audit"
	infraMLC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/mlc"
	infraStore "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/store"
)

type Classifier interface {
	ClassifyReview(ctx context.Context, text string) (infraMLC.Classification, error)
	CompleteJSON(ctx context.Context, prompt string, maxTokens int) (string, error)
}

type Service struct {
	repo    *pgAudit.Repository
	store   *infraStore.Client
	scope   *llmScope.Resolver
	mlc     Classifier
}

func NewService(repo *pgAudit.Repository, store *infraStore.Client, scope *llmScope.Resolver, mlc Classifier) *Service {
	return &Service{repo: repo, store: store, scope: scope, mlc: mlc}
}

func (s *Service) SearchApps(ctx context.Context, query, country, lang string) (auditDTO.SearchAppsResponse, error) {
	if s.store == nil || !s.store.Available() {
		return auditDTO.SearchAppsResponse{}, errors.New("store worker is not configured")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return auditDTO.SearchAppsResponse{}, errors.New("query is required")
	}
	country = normalizeCountry(country)
	lang = normalizeLang(lang)
	_ = s.store.Warmup(ctx)
	var play []infraStore.AppResult
	var appstore []infraStore.AppResult
	var playErr, appstoreErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		play, playErr = s.store.Search(ctx, "play", query, country, lang, 8)
	}()
	go func() {
		defer wg.Done()
		appstore, appstoreErr = s.store.Search(ctx, "appstore", query, country, lang, 8)
	}()
	wg.Wait()
	if playErr != nil {
		return auditDTO.SearchAppsResponse{}, playErr
	}
	if appstoreErr != nil {
		return auditDTO.SearchAppsResponse{}, appstoreErr
	}
	return auditDTO.SearchAppsResponse{
		Play:     mapApps(play),
		AppStore: mapApps(appstore),
	}, nil
}

func (s *Service) CreateAudit(ctx context.Context, userID string, req auditDTO.CreateAuditRequest) (auditDTO.AuditResponse, error) {
	if s.store == nil || !s.store.Available() {
		return auditDTO.AuditResponse{}, errors.New("store worker is not configured")
	}
	if strings.TrimSpace(req.ClientName) == "" || strings.TrimSpace(req.AppDisplayName) == "" {
		return auditDTO.AuditResponse{}, errors.New("client_name and app_display_name are required")
	}
	if req.PlayAppID == "" && req.AppStoreAppID == "" {
		return auditDTO.AuditResponse{}, errors.New("select at least one store app")
	}
	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = auditModel.ModeQuick
	}
	if mode != auditModel.ModeQuick && mode != auditModel.ModeFull {
		return auditDTO.AuditResponse{}, errors.New("mode must be quick or full")
	}

	orgID := ""
	if s.scope != nil {
		if m := s.scope.ForUser(ctx, userID); m.OrgID != "" {
			orgID = m.OrgID
		}
	}

	playLimit := 0
	if strings.TrimSpace(req.PlayAppID) != "" {
		playLimit = resolveReviewLimit(mode, req.PlayReviewLimit)
	}
	appStoreLimit := 0
	if strings.TrimSpace(req.AppStoreAppID) != "" {
		appStoreLimit = resolveReviewLimit(mode, req.AppStoreReviewLimit)
	}

	audit, err := s.repo.Create(ctx, auditModel.Audit{
		UserID:              userID,
		OrgID:               orgID,
		ClientName:          strings.TrimSpace(req.ClientName),
		AppDisplayName:      strings.TrimSpace(req.AppDisplayName),
		PlayAppID:           strings.TrimSpace(req.PlayAppID),
		AppStoreAppID:       strings.TrimSpace(req.AppStoreAppID),
		Country:             normalizeCountry(req.Country),
		Lang:                normalizeLang(req.Lang),
		PlayReviewLimit:     playLimit,
		AppStoreReviewLimit: appStoreLimit,
		Mode:                mode,
	})
	if err != nil {
		return auditDTO.AuditResponse{}, errors.New("could not create audit")
	}

	go s.runAudit(audit.ID, userID, orgID)

	return toAuditResponse(audit), nil
}

func (s *Service) GetAudit(ctx context.Context, userID, auditID string) (auditDTO.AuditResponse, error) {
	orgID := s.orgID(ctx, userID)
	a, err := s.repo.Get(ctx, auditID, userID, orgID)
	if err != nil {
		if errors.Is(err, pgAudit.ErrNotFound) {
			return auditDTO.AuditResponse{}, errors.New("audit not found")
		}
		return auditDTO.AuditResponse{}, err
	}
	return toAuditResponse(a), nil
}

func (s *Service) ListAudits(ctx context.Context, userID string) ([]auditDTO.AuditResponse, error) {
	orgID := s.orgID(ctx, userID)
	list, err := s.repo.List(ctx, userID, orgID, 50)
	if err != nil {
		return nil, errors.New("could not list audits")
	}
	out := make([]auditDTO.AuditResponse, 0, len(list))
	for _, a := range list {
		out = append(out, toAuditResponse(a))
	}
	return out, nil
}

func (s *Service) DeleteAudit(ctx context.Context, userID, auditID string) error {
	orgID := s.orgID(ctx, userID)
	if err := s.repo.Delete(ctx, auditID, userID, orgID); err != nil {
		if errors.Is(err, pgAudit.ErrNotFound) {
			return errors.New("audit not found")
		}
		return errors.New("could not delete audit")
	}
	return nil
}

func (s *Service) GetReport(ctx context.Context, userID, auditID string) (auditDTO.ReportResponse, error) {
	orgID := s.orgID(ctx, userID)
	a, err := s.repo.Get(ctx, auditID, userID, orgID)
	if err != nil {
		if errors.Is(err, pgAudit.ErrNotFound) {
			return auditDTO.ReportResponse{}, errors.New("audit not found")
		}
		return auditDTO.ReportResponse{}, err
	}
	resp := auditDTO.ReportResponse{Audit: toAuditResponse(a)}
	ins, err := s.repo.GetInsights(ctx, auditID)
	if err != nil || ins == nil {
		return resp, nil
	}
	statsMap := map[string]any{}
	statsJSON, _ := json.Marshal(ins.Statistics)
	_ = json.Unmarshal(statsJSON, &statsMap)
	root := make([]map[string]any, 0, len(ins.RootCauses))
	for _, rc := range ins.RootCauses {
		b, _ := json.Marshal(rc)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		root = append(root, m)
	}
	plan := make([]map[string]any, 0, len(ins.ActionPlan))
	for _, item := range ins.ActionPlan {
		b, _ := json.Marshal(item)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		plan = append(plan, m)
	}
	toMaps := func(v any) []map[string]any {
		b, _ := json.Marshal(v)
		var arr []map[string]any
		_ = json.Unmarshal(b, &arr)
		return arr
	}
	resp.Insights = &auditDTO.InsightsDTO{
		ExecutiveSummary:   ins.ExecutiveSummary,
		Statistics:         statsMap,
		RootCauses:         root,
		ActionPlan:         plan,
		CategoryInsights:   toMaps(ins.CategoryInsights),
		FeatureSuggestions: toMaps(ins.FeatureSuggestions),
		BugSuggestions:     toMaps(ins.BugSuggestions),
		FeaturedReviews:    toMaps(ins.FeaturedReviews),
		GeneratedAt:        ins.GeneratedAt.UTC().Format(time.RFC3339),
	}
	if ins.Statistics.ReportMeta != nil {
		b, _ := json.Marshal(ins.Statistics.ReportMeta)
		var meta map[string]any
		_ = json.Unmarshal(b, &meta)
		resp.Insights.ReportMeta = meta
	}
	return resp, nil
}

func (s *Service) runAudit(auditID, userID, orgID string) {
	ctx := context.Background()
	fail := func(msg string) {
		_ = s.repo.UpdateProgress(ctx, auditID, map[string]any{
			"status": auditModel.StatusFailed, "step": auditModel.StepDone, "error_message": msg,
			"play_fetched": -1, "appstore_fetched": -1, "total_reviews": -1, "analyzed_count": -1,
		})
	}

	_ = s.repo.UpdateProgress(ctx, auditID, map[string]any{
		"status": auditModel.StatusRunning, "step": auditModel.StepCrawl,
		"play_fetched": -1, "appstore_fetched": -1, "total_reviews": -1, "analyzed_count": -1,
	})

	a, err := s.repo.Get(ctx, auditID, userID, orgID)
	if err != nil {
		fail("audit not found")
		return
	}

	limit := reviewLimit(a.Mode)
	playLimit := a.PlayReviewLimit
	if playLimit <= 0 {
		playLimit = limit
	}
	appStoreLimit := a.AppStoreReviewLimit
	if appStoreLimit <= 0 {
		appStoreLimit = limit
	}
	_ = s.store.Warmup(ctx)
	crawl, err := s.store.CrawlApps(ctx, infraStore.CrawlOptions{
		AppName:             a.AppDisplayName,
		PlayAppID:           a.PlayAppID,
		AppStoreAppID:       a.AppStoreAppID,
		PlayReviewLimit:     playLimit,
		AppStoreReviewLimit: appStoreLimit,
		Lang:                a.Lang,
		Country:             a.Country,
	})
	if err != nil {
		fail(err.Error())
		return
	}

	rows := make([]auditModel.Review, 0, len(crawl.Reviews))
	for _, r := range crawl.Reviews {
		if strings.TrimSpace(r.Text) == "" {
			continue
		}
		var reviewedAt *time.Time
		if r.ReviewedAt != "" {
			if t, parseErr := time.Parse(time.RFC3339, r.ReviewedAt); parseErr == nil {
				reviewedAt = &t
			}
		}
		rows = append(rows, auditModel.Review{
			AuditID:       auditID,
			Store:         r.Store,
			StoreReviewID: r.StoreReviewID,
			AppName:       r.AppName,
			Rating:        r.Rating,
			Text:          r.Text,
			ReviewedAt:    reviewedAt,
		})
	}
	if _, err := s.repo.InsertReviews(ctx, auditID, rows); err != nil {
		fail("could not save reviews")
		return
	}

	_ = s.repo.UpdateProgress(ctx, auditID, map[string]any{
		"step":             auditModel.StepClassify,
		"play_fetched":     crawl.PlayCount,
		"appstore_fetched": crawl.AppStoreCount,
		"total_reviews":    len(rows),
		"truncated":        crawl.Truncated,
		"analyzed_count":   0,
	})

	if s.mlc == nil {
		fail("mlc inference is not configured")
		return
	}

	analyzed, err := s.batchClassify(ctx, auditID)
	if err != nil {
		fail(err.Error())
		return
	}

	_ = s.repo.UpdateProgress(ctx, auditID, map[string]any{
		"step": auditModel.StepInsights, "analyzed_count": analyzed,
		"play_fetched": -1, "appstore_fetched": -1, "total_reviews": -1,
	})

	stats, err := s.repo.ComputeStatistics(ctx, auditID)
	if err != nil {
		fail("could not compute statistics")
		return
	}
	stats = enrichStatistics(stats)
	samples, _ := s.repo.SampleNegativeReviews(ctx, auditID, 60)
	vertical := detectAppVertical(a.AppDisplayName, a.ClientName, samples)
	stats = enrichStatisticsForVertical(stats, vertical)
	if err := s.generateInsights(ctx, auditID, a, stats, samples, vertical); err != nil {
		fail(err.Error())
		return
	}

	_ = s.repo.UpdateProgress(ctx, auditID, map[string]any{
		"status": auditModel.StatusCompleted, "step": auditModel.StepDone,
		"play_fetched": -1, "appstore_fetched": -1, "total_reviews": -1, "analyzed_count": analyzed,
	})
}

func (s *Service) batchClassify(ctx context.Context, auditID string) (int, error) {
	batchSize := envInt("AUDIT_BATCH_SIZE", 8)
	workers := envInt("AUDIT_CLASSIFY_WORKERS", 4)
	total := 0

	for {
		pending, err := s.repo.ListUnclassified(ctx, auditID, batchSize*workers*2)
		if err != nil {
			return total, err
		}
		if len(pending) == 0 {
			return total, nil
		}

		type job struct {
			items []infraMLC.BatchReviewInput
			reviews []auditModel.Review
		}
		jobs := []job{}
		for i := 0; i < len(pending); i += batchSize {
			end := i + batchSize
			if end > len(pending) {
				end = len(pending)
			}
			chunk := pending[i:end]
			inputs := make([]infraMLC.BatchReviewInput, len(chunk))
			for j, rv := range chunk {
				inputs[j] = infraMLC.BatchReviewInput{ID: rv.ID, Text: rv.Text}
			}
			jobs = append(jobs, job{items: inputs, reviews: chunk})
		}

		var wg sync.WaitGroup
		errCh := make(chan error, len(jobs))
		sem := make(chan struct{}, workers)
		var mu sync.Mutex

		for _, jb := range jobs {
			wg.Add(1)
			go func(jb job) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				for _, rv := range jb.reviews {
					cls, err := s.mlc.ClassifyReview(ctx, rv.Text)
					if err != nil {
						errCh <- err
						return
					}
					if err := s.repo.UpdateReviewClassification(ctx, rv.ID, cls.Category, cls.Sentiment, cls.RawOutput); err != nil {
						errCh <- err
						return
					}
					mu.Lock()
					total++
					cur := total
					mu.Unlock()
					_ = s.repo.UpdateProgress(ctx, auditID, map[string]any{
						"analyzed_count": cur, "play_fetched": -1, "appstore_fetched": -1, "total_reviews": -1,
					})
				}
			}(jb)
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				return total, err
			}
		}
	}
}

func (s *Service) generateInsights(ctx context.Context, auditID string, a auditModel.Audit, stats auditModel.Statistics, samples []auditModel.Review, vertical appVertical) error {
	statsJSON, _ := json.Marshal(stats)
	var sampleLines strings.Builder
	for i, rv := range samples {
		if i >= 40 {
			break
		}
		sampleLines.WriteString(fmt.Sprintf("- [%s %d★ %s/%s] %q\n", rv.Store, rv.Rating, rv.Category, rv.Sentiment, truncate(rv.Text, 180)))
	}

	baseMeta := fallbackReportMeta(stats, fallbackRootCauses(stats, samples, vertical), a.AppDisplayName, vertical)
	baseMeta.Scenarios = nil

	classified, _ := s.repo.ListClassifiedReviews(ctx, auditID, 800)
	categoryInsights := buildCategoryInsights(stats, classified, vertical)
	featureSuggestions := buildFeatureSuggestions(classified)
	bugSuggestions := buildBugSuggestions(classified)
	featuredReviews := buildFeaturedReviews(classified)
	stats.StoreBreakdown = buildStoreBreakdown(a, classified, vertical)

	prompt := fmt.Sprintf(`Sen mobil uygulama danışmanısın. Aşağıdaki uygulama için Türkçe müşteri sunumu raporu üret.

Uygulama: "%s"
Müşteri/marka: "%s"
Sektör/vertical: %s
Mevcut yazılı yorum ortalaması: %.2f

ZORUNLU kurallar:
- Sadece verilen yorum örneklerine dayan; uydurma özellik veya hata yazma.
- feature_suggestions: kullanıcıların istediği özellikler (ör. "gemi sistemi olsun") — her biri supporting_reviews ile yorum alıntısı içersin.
- bug_suggestions: crash, performans, monetization şikâyetleri — yorum alıntılarıyla destekle.
- category_insights: bug/feature/praise/other kategorileri ve her biri için örnek yorum alıntıları.
- featured_reviews: en dikkat çekici 4-6 yorum (olumlu + olumsuz karışık).
- scenarios YAZMA.
- timeline ve action_plan sektöre uygun olsun (%s).

SADECE geçerli JSON döndür:
{
  "executive_summary": "max 180 kelime",
  "callout": "tek paragraf",
  "root_causes": [{"theme":"","description":"","affected_rating":"1-3","sample_count":0,"examples":["",""]}],
  "category_insights": [{"category":"bug|feature|praise|other","label":"","count":0,"pct":0,"reviews":[{"text":"","rating":1,"store":"play|appstore","sentiment":""}]}],
  "feature_suggestions": [{"title":"","summary":"","category":"feature","priority":"P1","supporting_reviews":[{"text":"","rating":4,"store":"","sentiment":""}]}],
  "bug_suggestions": [{"title":"","summary":"","category":"bug","priority":"P0","supporting_reviews":[{"text":"","rating":1,"store":"","sentiment":""}]}],
  "featured_reviews": [{"text":"","rating":5,"store":"","category":"","sentiment":"","highlight":""}],
  "timeline": [{"horizon":"0-30 gün","tag":"fast|mid|hard","title":"","body":""}],
  "priorities": [{"rank":1,"title":"","body":""}],
  "management_findings": [{"title":"","body":""}],
  "action_plan": [{"priority":"P0|P1|P2","horizon_days":"","title":"","action":"","owner_hint":"","expected_impact":"","tag":""}]
}

İstatistikler: %s
Sınıflandırılmış yorum örnekleri:
%s`, a.AppDisplayName, a.ClientName, verticalLabel(vertical), stats.AvgRating,
		verticalLabel(vertical), string(statsJSON), sampleLines.String())

	raw, err := s.mlc.CompleteJSON(ctx, prompt, 2800)
	rootCauses := fallbackRootCauses(stats, samples, vertical)
	actionPlan := fallbackActionPlan(baseMeta)
	meta := baseMeta
	summary := buildFallbackSummary(a, stats, vertical)

	if err == nil {
		if llmMeta, llmRoot, llmPlan, llmSummary, parseErr := parseInsightBundle(raw); parseErr == nil {
			llmMeta = sanitizeReportMeta(llmMeta, vertical, a.AppDisplayName)
			llmMeta.Scenarios = nil
			meta = mergeReportMeta(baseMeta, llmMeta, stats)
			meta.Scenarios = nil
			if len(llmRoot) > 0 {
				rootCauses = sanitizeRootCauses(llmRoot, vertical)
			}
			if len(llmPlan) > 0 {
				actionPlan = sanitizeActionPlan(llmPlan, vertical)
			}
			if strings.TrimSpace(llmSummary) != "" && !mentionsForeignBrand(llmSummary, a.AppDisplayName, a.ClientName) {
				summary = strings.TrimSpace(llmSummary)
			}
		} else {
			if rc := parseRootCauses(raw); len(rc) > 0 {
				rootCauses = sanitizeRootCauses(rc, vertical)
			}
			if ap := parseActionPlan(raw); len(ap) > 0 {
				actionPlan = sanitizeActionPlan(ap, vertical)
			}
		}
		llmCats, llmFeat, llmBug, llmFeatured := parseFeedbackFromRaw(raw)
		categoryInsights = mergeCategoryInsights(categoryInsights, llmCats)
		categoryInsights = applyVerticalCategoryLabels(categoryInsights, vertical)
		featureSuggestions = mergeFeedbackSuggestions(featureSuggestions, llmFeat, "feature")
		bugSuggestions = mergeFeedbackSuggestions(bugSuggestions, llmBug, "bug")
		featuredReviews = mergeFeaturedReviews(featuredReviews, llmFeatured)
	}

	meta.CurrentAvgRating = stats.AvgRating
	meta.StretchGoalRating = stretchGoalRating(stats.AvgRating)
	stats.ReportMeta = &meta

	return s.repo.SaveInsights(ctx, auditModel.Insights{
		AuditID:            auditID,
		ExecutiveSummary:   summary,
		Statistics:         stats,
		RootCauses:         rootCauses,
		ActionPlan:         actionPlan,
		CategoryInsights:   categoryInsights,
		FeatureSuggestions: featureSuggestions,
		BugSuggestions:     bugSuggestions,
		FeaturedReviews:    featuredReviews,
	})
}

func parseRootCauses(raw string) []auditModel.RootCause {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil
	}
	var payload struct {
		RootCauses []auditModel.RootCause `json:"root_causes"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &payload); err != nil {
		return nil
	}
	return payload.RootCauses
}

func parseActionPlan(raw string) []auditModel.ActionItem {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil
	}
	var payload struct {
		ActionPlan []auditModel.ActionItem `json:"action_plan"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &payload); err != nil {
		return nil
	}
	return payload.ActionPlan
}

func mapApps(in []infraStore.AppResult) []auditDTO.StoreApp {
	out := make([]auditDTO.StoreApp, 0, len(in))
	for _, a := range in {
		out = append(out, auditDTO.StoreApp{
			Store: a.Store, AppID: a.AppID, AppName: a.AppName, Developer: a.Developer, IconURL: a.IconURL,
		})
	}
	return out
}

func toAuditResponse(a auditModel.Audit) auditDTO.AuditResponse {
	resp := auditDTO.AuditResponse{
		ID: a.ID, ClientName: a.ClientName, AppDisplayName: a.AppDisplayName,
		PlayAppID: a.PlayAppID, AppStoreAppID: a.AppStoreAppID,
		Country: a.Country, Lang: a.Lang,
		PlayReviewLimit: a.PlayReviewLimit, AppStoreReviewLimit: a.AppStoreReviewLimit,
		Mode: a.Mode,
		Status: a.Status, Step: a.Step, PlayFetched: a.PlayFetched, AppStoreFetched: a.AppStoreFetched,
		TotalReviews: a.TotalReviews, AnalyzedCount: a.AnalyzedCount, Truncated: a.Truncated,
		ErrorMessage: a.ErrorMessage, CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339),
	}
	if a.StartedAt != nil {
		resp.StartedAt = a.StartedAt.UTC().Format(time.RFC3339)
	}
	if a.CompletedAt != nil {
		resp.CompletedAt = a.CompletedAt.UTC().Format(time.RFC3339)
	}
	return resp
}

func (s *Service) orgID(ctx context.Context, userID string) string {
	if s.scope == nil {
		return ""
	}
	return s.scope.ForUser(ctx, userID).OrgID
}

func reviewLimit(mode string) int {
	if mode == auditModel.ModeFull {
		return envInt("AUDIT_MAX_REVIEWS", 10000)
	}
	return envInt("AUDIT_QUICK_REVIEW_CAP", 500)
}

func resolveReviewLimit(mode string, requested int) int {
	if requested > 0 {
		return requested
	}
	return reviewLimit(mode)
}

func normalizeCountry(country string) string {
	country = strings.ToLower(strings.TrimSpace(country))
	if country == "" {
		return "tr"
	}
	return country
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		return "tr"
	}
	return lang
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
