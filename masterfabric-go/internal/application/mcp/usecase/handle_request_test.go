package usecase_test

import (
	"context"
	"testing"

	mcpUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/mcp/usecase"
	mcpModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/mcp/model"
)

func TestHandleMCPRequest_UnsupportedAction(t *testing.T) {
	_, err := mcpUC.HandleMCPRequest(context.Background(), "user-1", mcpModel.Payload{
		Action: "deepkwiki_search",
	}, nil)
	if err == nil || err.Error() != "unsupported mcp action" {
		t.Fatalf("expected unsupported mcp action, got %v", err)
	}
}

func TestHandleMCPRequest_NilAnalyze(t *testing.T) {
	_, err := mcpUC.HandleMCPRequest(context.Background(), "user-1", mcpModel.Payload{
		Action:   "analyze_review",
		ReviewID: "rev-1",
	}, nil)
	if err == nil || err.Error() != "mlc inference is not configured" {
		t.Fatalf("expected mlc inference is not configured, got %v", err)
	}
}
