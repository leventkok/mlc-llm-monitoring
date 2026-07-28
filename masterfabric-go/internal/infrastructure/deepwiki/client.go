package deepwiki

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultMCPURL = "https://mcp.deepwiki.com/mcp"

// Client calls the public DeepWiki MCP HTTP endpoint (no auth).
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultMCPURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type rpcResponse struct {
	Result *struct {
		StructuredContent *struct {
			Result string `json:"result"`
		} `json:"structuredContent"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"result"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) AskQuestion(ctx context.Context, repoName, question string) (string, error) {
	return c.callTool(ctx, "ask_question", map[string]any{
		"repoName": repoName,
		"question": question,
	})
}

func (c *Client) ReadWikiStructure(ctx context.Context, repoName string) (string, error) {
	return c.callTool(ctx, "read_wiki_structure", map[string]any{
		"repoName": repoName,
	})
}

func (c *Client) ReadWikiContents(ctx context.Context, repoName string) (string, error) {
	return c.callTool(ctx, "read_wiki_contents", map[string]any{
		"repoName": repoName,
	})
}

func (c *Client) callTool(ctx context.Context, tool string, args map[string]any) (string, error) {
	body, err := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: toolCallParams{
			Name:      tool,
			Arguments: args,
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("deepwiki request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("deepwiki returned %d: %s", resp.StatusCode, string(raw))
	}

	text, err := parseSSEResponse(string(raw))
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(text, "Error fetching wiki") {
		return text, fmt.Errorf("%s", text)
	}
	return text, nil
}

func parseSSEResponse(raw string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		var msg rpcResponse
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			continue
		}
		if msg.Error != nil {
			return "", fmt.Errorf("deepwiki mcp: %s", msg.Error.Message)
		}
		if msg.Result == nil {
			continue
		}
		if msg.Result.StructuredContent != nil && msg.Result.StructuredContent.Result != "" {
			return msg.Result.StructuredContent.Result, nil
		}
		if len(msg.Result.Content) > 0 && msg.Result.Content[0].Text != "" {
			return msg.Result.Content[0].Text, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("deepwiki mcp: empty response")
}
