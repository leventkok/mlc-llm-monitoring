package usecase_test

import (
	"context"
	"testing"

	mcpUC "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/application/mcp/usecase"
	llmModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/llm/model"
	mcpModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/mcp/model"
)

func userScope(userID string) llmModel.ReviewScope {
	return llmModel.ReviewScope{UserID: userID}
}

func TestHandleMCPRequest_UnsupportedAction(t *testing.T) {
	_, err := mcpUC.HandleMCPRequest(context.Background(), userScope("user-1"), mcpModel.Payload{
		Action: "unknown",
	}, nil, nil, nil)
	if err == nil || err.Error() != "unsupported mcp action" {
		t.Fatalf("expected unsupported mcp action, got %v", err)
	}
}

func TestHandleMCPRequest_NilAnalyze(t *testing.T) {
	_, err := mcpUC.HandleMCPRequest(context.Background(), userScope("user-1"), mcpModel.Payload{
		Action:   "analyze_review",
		ReviewID: "rev-1",
	}, nil, nil, nil)
	if err == nil || err.Error() != "mlc inference is not configured" {
		t.Fatalf("expected mlc inference is not configured, got %v", err)
	}
}

func TestHandleDeepKwiki_RequiresRepo(t *testing.T) {
	_, err := mcpUC.HandleMCPRequest(context.Background(), userScope("user-1"), mcpModel.Payload{
		Action: "deepkwiki_search",
		Query:  "what is this repo?",
	}, nil, nil, nil)
	if err == nil || err.Error() != "spec.repo is required (owner/name)" {
		t.Fatalf("expected spec.repo is required, got %v", err)
	}
}

type stubDeepWiki struct {
	answer string
}

func (s stubDeepWiki) AskQuestion(_ context.Context, _, _ string) (string, error) {
	return s.answer, nil
}

func (s stubDeepWiki) ReadWikiStructure(_ context.Context, _ string) (string, error) {
	return "structure", nil
}

func (s stubDeepWiki) ReadWikiContents(_ context.Context, _ string) (string, error) {
	return "contents", nil
}

func TestHandleDeepKwiki_Ask(t *testing.T) {
	result, err := mcpUC.HandleMCPRequest(context.Background(), userScope("user-1"), mcpModel.Payload{
		Action: "deepkwiki_search",
		Query:  "What is MLC?",
		Spec:   map[string]any{"repo": "mlc-ai/mlc-llm", "mode": "ask"},
	}, nil, stubDeepWiki{answer: "MLC is a compiler."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "deepkwiki_search" || result.Answer != "MLC is a compiler." {
		t.Fatalf("unexpected result: %+v", result)
	}
}
