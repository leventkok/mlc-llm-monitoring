package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	llmUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/usecase"
	mcpModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/mcp/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/deepwiki"
)

// DeepWikiQuerier fetches repo documentation from DeepWiki MCP.
type DeepWikiQuerier interface {
	AskQuestion(ctx context.Context, repoName, question string) (string, error)
	ReadWikiStructure(ctx context.Context, repoName string) (string, error)
	ReadWikiContents(ctx context.Context, repoName string) (string, error)
}

// HandleMCPRequest validates MCP payloads and routes to the appropriate handler.
func HandleMCPRequest(
	ctx context.Context,
	userID string,
	req mcpModel.Payload,
	analyze *llmUC.AnalyzeReviewUseCase,
	deepWiki DeepWikiQuerier,
) (mcpModel.RichResult, error) {
	switch req.Action {
	case "analyze_review":
		return handleAnalyzeReview(ctx, userID, req, analyze)
	case "deepkwiki_search":
		return handleDeepKwikiSearch(ctx, req, deepWiki)
	default:
		return mcpModel.RichResult{}, errors.New("unsupported mcp action")
	}
}

func handleAnalyzeReview(
	ctx context.Context,
	userID string,
	req mcpModel.Payload,
	analyze *llmUC.AnalyzeReviewUseCase,
) (mcpModel.RichResult, error) {
	if analyze == nil {
		return mcpModel.RichResult{}, errors.New("mlc inference is not configured")
	}
	if req.ReviewID == "" {
		return mcpModel.RichResult{}, errors.New("review_id is required")
	}

	decision, err := analyze.Execute(ctx, userID, req.ReviewID)
	if err != nil {
		return mcpModel.RichResult{}, err
	}

	return mcpModel.RichResult{
		Action:     "analyze_review",
		ReviewID:   req.ReviewID,
		DecisionID: decision.ID,
		Category:   decision.Category,
		Sentiment:  decision.Sentiment,
		RawOutput:  decision.RawOutput,
		LatencyMs:  decision.LatencyMs,
	}, nil
}

func handleDeepKwikiSearch(
	ctx context.Context,
	req mcpModel.Payload,
	deepWiki DeepWikiQuerier,
) (mcpModel.RichResult, error) {
	repo := repoFromSpec(req.Spec)
	if repo == "" {
		return mcpModel.RichResult{}, errors.New("spec.repo is required (owner/name)")
	}
	if deepWiki == nil {
		return mcpModel.RichResult{}, errors.New("deepwiki is not configured")
	}

	mode := modeFromSpec(req.Spec)
	query := strings.TrimSpace(req.Query)
	start := time.Now()

	var (
		answer string
		err    error
	)
	switch mode {
	case "structure":
		answer, err = deepWiki.ReadWikiStructure(ctx, repo)
	case "contents":
		answer, err = deepWiki.ReadWikiContents(ctx, repo)
	default:
		if query == "" {
			return mcpModel.RichResult{}, errors.New("query is required for ask mode")
		}
		answer, err = deepWiki.AskQuestion(ctx, repo, query)
		mode = "ask"
	}
	if err != nil {
		return mcpModel.RichResult{}, err
	}

	return mcpModel.RichResult{
		Action:    "deepkwiki_search",
		RepoName:  repo,
		Query:     query,
		Mode:      mode,
		Answer:    answer,
		LatencyMs: int(time.Since(start).Milliseconds()),
	}, nil
}

func repoFromSpec(spec map[string]any) string {
	if spec == nil {
		return ""
	}
	for _, key := range []string{"repo", "repo_name", "repoName"} {
		if v, ok := spec[key].(string); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func modeFromSpec(spec map[string]any) string {
	if spec == nil {
		return "ask"
	}
	if v, ok := spec["mode"].(string); ok {
		return strings.TrimSpace(strings.ToLower(v))
	}
	return "ask"
}

// Ensure deepwiki.Client satisfies DeepWikiQuerier.
var _ DeepWikiQuerier = (*deepwiki.Client)(nil)
