package usecase

import (
	"context"
	"errors"

	llmUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/llm/usecase"
	mcpModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/mcp/model"
)

// HandleMCPRequest validates MCP payloads and routes to the appropriate handler.
func HandleMCPRequest(
	ctx context.Context,
	userID string,
	req mcpModel.Payload,
	analyze *llmUC.AnalyzeReviewUseCase,
) (mcpModel.RichResult, error) {
	switch req.Action {
	case "analyze_review":
		return handleAnalyzeReview(ctx, userID, req, analyze)
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
		ReviewID:   req.ReviewID,
		DecisionID: decision.ID,
		Category:   decision.Category,
		Sentiment:  decision.Sentiment,
		RawOutput:  decision.RawOutput,
		LatencyMs:  decision.LatencyMs,
	}, nil
}
