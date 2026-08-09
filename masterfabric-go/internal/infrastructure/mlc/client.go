package mlc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/config/model"
	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/shared/validate"
)

const defaultBaseURL = "http://mlc-llm:8080"

var (
	categories = []string{"bug", "feature", "praise", "spam", "other"}
	sentiments = []string{"positive", "negative", "neutral"}
)

// LLMSettingsReader supplies runtime LLM parameters (admin panel).
type LLMSettingsReader interface {
	GetLLM() model.LLMConfig
}

// Classification is the parsed LLM output for a review.
type Classification struct {
	Category  string
	Sentiment string
	RawOutput string
	LatencyMs int
}

// Client calls an OpenAI-compatible MLC LLM HTTP API.
type Client struct {
	baseURL    string
	model      string
	apiKey     string
	settings   LLMSettingsReader
	httpClient *http.Client
}

func NewClient(baseURL, model, apiKey string, settings LLMSettingsReader) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if model == "" {
		model = "gemma-2-2b-it-q4f16_1-MLC"
	}
	return &Client{
		baseURL:  baseURL,
		model:    model,
		apiKey:   strings.TrimSpace(apiKey),
		settings: settings,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-MLC-API-Key", c.apiKey)
	}
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	TopP        float64       `json:"top_p,omitempty"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) llmSettings() model.LLMConfig {
	if c.settings != nil {
		return c.settings.GetLLM()
	}
	return model.DefaultLLMConfig(c.model, "")
}

func (c *Client) buildPrompt(text string) string {
	s := c.llmSettings()
	base := strings.TrimSpace(s.SystemPrompt)
	if base == "" {
		base = model.DefaultSystemPrompt
	}
	return base + "\n\nReview: " + fmt.Sprintf("%q", text)
}

func (c *Client) ClassifyReview(ctx context.Context, text string) (Classification, error) {
	start := time.Now()
	s := c.llmSettings()
	modelID := c.model
	if m := strings.TrimSpace(s.ActiveModel); m != "" {
		modelID = m
	}
	prompt := c.buildPrompt(text)

	body, err := json.Marshal(chatRequest{
		Model: modelID,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: s.Temperature,
		TopP:        s.TopP,
		MaxTokens:   s.MaxTokens,
	})
	if err != nil {
		return Classification{}, fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Classification{}, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Classification{}, fmt.Errorf("mlc request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Classification{}, fmt.Errorf("read mlc response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return Classification{}, fmt.Errorf("mlc returned %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return Classification{}, fmt.Errorf("decode mlc response: %w", err)
	}

	raw := ""
	if len(parsed.Choices) > 0 {
		raw = strings.TrimSpace(parsed.Choices[0].Message.Content)
	}

	result := Classification{
		Category:  "other",
		Sentiment: "neutral",
		RawOutput: raw,
		LatencyMs: int(time.Since(start).Milliseconds()),
	}

	if idx := strings.Index(raw, "{"); idx >= 0 {
		if end := strings.LastIndex(raw, "}"); end > idx {
			var obj struct {
				Category  string `json:"category"`
				Sentiment string `json:"sentiment"`
			}
			if err := json.Unmarshal([]byte(raw[idx:end+1]), &obj); err == nil {
				if err := validate.Category(obj.Category); err == nil {
					result.Category = obj.Category
				}
				if err := validate.Sentiment(obj.Sentiment); err == nil {
					result.Sentiment = obj.Sentiment
				}
			}
		}
	}

	return result, nil
}

type BatchReviewInput struct {
	ID   string
	Text string
}

type BatchClassification struct {
	ID        string
	Category  string
	Sentiment string
	RawOutput string
}

func (c *Client) ClassifyReviewBatch(ctx context.Context, items []BatchReviewInput) ([]BatchClassification, error) {
	if len(items) == 0 {
		return nil, nil
	}
	s := c.llmSettings()
	modelID := c.model
	if m := strings.TrimSpace(s.ActiveModel); m != "" {
		modelID = m
	}
	var b strings.Builder
	b.WriteString("Classify each app store review. Return ONLY a JSON array with one object per review in order.\n")
	b.WriteString("Each object: {\"id\":\"...\",\"category\":\"bug|feature|praise|spam|other\",\"sentiment\":\"positive|negative|neutral\"}\n\n")
	for i, item := range items {
		b.WriteString(fmt.Sprintf("%d) id=%s review=%q\n", i+1, item.ID, item.Text))
	}
	raw, err := c.complete(ctx, modelID, b.String(), s.Temperature, s.TopP, maxInt(s.MaxTokens, 512))
	if err != nil {
		return nil, err
	}
	return parseBatchClassifications(raw, items), nil
}

func (c *Client) CompleteJSON(ctx context.Context, prompt string, maxTokens int) (string, error) {
	s := c.llmSettings()
	modelID := c.model
	if m := strings.TrimSpace(s.ActiveModel); m != "" {
		modelID = m
	}
	if maxTokens <= 0 {
		maxTokens = 1200
	}
	return c.complete(ctx, modelID, prompt, s.Temperature, s.TopP, maxTokens)
}

func (c *Client) complete(ctx context.Context, modelID, prompt string, temp, topP float64, maxTokens int) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model: modelID,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: temp,
		TopP:        topP,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("mlc request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("mlc returned %d: %s", resp.StatusCode, string(respBody))
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty mlc response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func parseBatchClassifications(raw string, items []BatchReviewInput) []BatchClassification {
	out := make([]BatchClassification, len(items))
	for i, item := range items {
		out[i] = BatchClassification{ID: item.ID, Category: "other", Sentiment: "neutral", RawOutput: raw}
	}
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start < 0 || end <= start {
		return out
	}
	var arr []struct {
		ID        string `json:"id"`
		Category  string `json:"category"`
		Sentiment string `json:"sentiment"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &arr); err != nil {
		return out
	}
	byID := map[string]BatchClassification{}
	for _, row := range arr {
		cat, sent := "other", "neutral"
		if validate.Category(row.Category) == nil {
			cat = row.Category
		}
		if validate.Sentiment(row.Sentiment) == nil {
			sent = row.Sentiment
		}
		byID[row.ID] = BatchClassification{ID: row.ID, Category: cat, Sentiment: sent, RawOutput: raw}
	}
	for i, item := range items {
		if v, ok := byID[item.ID]; ok {
			out[i] = v
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mlc health status %d", resp.StatusCode)
	}
	return nil
}
