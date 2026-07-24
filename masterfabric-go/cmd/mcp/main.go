package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	loadEnv()

	client, err := newAPIClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "inferreview MCP: %v\n", err)
		os.Exit(1)
	}

	s := server.NewMCPServer(
		"inferreview",
		"0.1.0",
		server.WithToolCapabilities(true),
	)

	registerTools(s, client)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "inferreview MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func registerTools(s *server.MCPServer, client *apiClient) {
	s.AddTool(mcp.NewTool("login",
		mcp.WithDescription("Sign in to the inferreview API (stores session cookie for this MCP process)"),
		mcp.WithString("email", mcp.Required(), mcp.Description("Account email")),
		mcp.WithString("password", mcp.Required(), mcp.Description("Account password")),
	), toolHandler(client, func(c *apiClient, args map[string]any) ([]byte, error) {
		email, _ := args["email"].(string)
		password, _ := args["password"].(string)
		if err := c.login(email, password); err != nil {
			return nil, err
		}
		me, err := c.get("/auth/me")
		if err != nil {
			return nil, err
		}
		return me, nil
	}))

	s.AddTool(mcp.NewTool("get_health",
		mcp.WithDescription("Check API health (no auth required)"),
	), toolHandler(client, func(c *apiClient, _ map[string]any) ([]byte, error) {
		return c.do("GET", "/health", nil, false)
	}))

	s.AddTool(mcp.NewTool("get_stats",
		mcp.WithDescription("Get monitoring KPI stats for the signed-in user (/stats)"),
	), toolHandler(client, func(c *apiClient, _ map[string]any) ([]byte, error) {
		return c.get("/stats")
	}))

	s.AddTool(mcp.NewTool("list_reviews",
		mcp.WithDescription("List app store reviews for the signed-in user"),
		mcp.WithNumber("limit", mcp.Description("Max rows (default 50)"), mcp.DefaultNumber(50)),
	), toolHandler(client, func(c *apiClient, args map[string]any) ([]byte, error) {
		limit := intArg(args, "limit", 50)
		return c.get(fmt.Sprintf("/reviews?limit=%d&offset=0", limit))
	}))

	s.AddTool(mcp.NewTool("create_review",
		mcp.WithDescription("Create a new app store review"),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Application name")),
		mcp.WithString("store", mcp.Description("play or appstore"), mcp.DefaultString("play")),
		mcp.WithNumber("rating", mcp.Required(), mcp.Description("1-5 star rating"), mcp.Min(1), mcp.Max(5)),
		mcp.WithString("text", mcp.Required(), mcp.Description("Review text")),
	), toolHandler(client, func(c *apiClient, args map[string]any) ([]byte, error) {
		payload := map[string]any{
			"app_name": args["app_name"],
			"store":    stringArg(args, "store", "play"),
			"rating":   intArg(args, "rating", 3),
			"text":     args["text"],
		}
		return c.post("/reviews", payload)
	}))

	s.AddTool(mcp.NewTool("analyze_review",
		mcp.WithDescription("Run MLC classification on a review (server-side analyze)"),
		mcp.WithString("review_id", mcp.Required(), mcp.Description("Review UUID")),
	), toolHandler(client, func(c *apiClient, args map[string]any) ([]byte, error) {
		id, _ := args["review_id"].(string)
		return c.post(fmt.Sprintf("/reviews/%s/analyze", id), nil)
	}))

	s.AddTool(mcp.NewTool("list_decisions",
		mcp.WithDescription("List LLM classification decisions"),
		mcp.WithNumber("limit", mcp.Description("Max rows (default 50)"), mcp.DefaultNumber(50)),
	), toolHandler(client, func(c *apiClient, args map[string]any) ([]byte, error) {
		limit := intArg(args, "limit", 50)
		return c.get(fmt.Sprintf("/decisions?limit=%d&offset=0", limit))
	}))

	s.AddTool(mcp.NewTool("list_scores",
		mcp.WithDescription("List auto quality scores for decisions"),
		mcp.WithNumber("limit", mcp.Description("Max rows (default 50)"), mcp.DefaultNumber(50)),
	), toolHandler(client, func(c *apiClient, args map[string]any) ([]byte, error) {
		limit := intArg(args, "limit", 50)
		return c.get(fmt.Sprintf("/scores?limit=%d&offset=0", limit))
	}))
}

type apiCall func(*apiClient, map[string]any) ([]byte, error)

func toolHandler(client *apiClient, call apiCall) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		raw, err := call(client, args)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text, err := prettyJSON(raw)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	}
}

func stringArg(args map[string]any, key, fallback string) string {
	if v, ok := args[key].(string); ok && v != "" {
		return v
	}
	return fallback
}

func intArg(args map[string]any, key string, fallback int) int {
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return fallback
	}
}

func loadEnv() {
	candidates := []string{
		filepath.Join("..", ".cursor", "mcp.env"),
		filepath.Join(".cursor", "mcp.env"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			return
		}
	}
}
