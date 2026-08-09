package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultWorkerURL = "http://store-worker:8091"

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultWorkerURL
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *Client) Available() bool {
	return c.baseURL != ""
}

type AppResult struct {
	Store     string `json:"store"`
	AppID     string `json:"app_id"`
	AppName   string `json:"app_name"`
	Developer string `json:"developer"`
	IconURL   string `json:"icon_url"`
}

type ReviewRow struct {
	Store         string `json:"store"`
	StoreReviewID string `json:"store_review_id"`
	AppName       string `json:"app_name"`
	Rating        int    `json:"rating"`
	Text          string `json:"text"`
	ReviewedAt    string `json:"reviewed_at"`
	Language      string `json:"language"`
}

type CrawlResult struct {
	AppName         string      `json:"app_name"`
	PlayAppID       string      `json:"play_app_id"`
	AppStoreAppID   string      `json:"appstore_app_id"`
	PlayCount       int         `json:"play_count"`
	AppStoreCount   int         `json:"appstore_count"`
	TotalReturned   int         `json:"total_returned"`
	Truncated       bool        `json:"truncated"`
	Reviews         []ReviewRow `json:"reviews"`
}

type crawlRequest struct {
	AppName        string `json:"app_name"`
	PlayAppID      string `json:"play_app_id"`
	AppStoreAppID  string `json:"appstore_app_id"`
	Limit          int    `json:"limit"`
	Lang           string `json:"lang"`
	Country        string `json:"country"`
}

func (c *Client) Search(ctx context.Context, store, query string, limit int) ([]AppResult, error) {
	u, err := url.Parse(c.baseURL + "/search")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("store", store)
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("store worker unreachable: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("store search %d: %s", resp.StatusCode, string(body))
	}
	var parsed struct {
		Apps []AppResult `json:"apps"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return parsed.Apps, nil
}

func (c *Client) Crawl(ctx context.Context, req crawlRequest) (CrawlResult, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return CrawlResult{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/crawl", bytes.NewReader(payload))
	if err != nil {
		return CrawlResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return CrawlResult{}, fmt.Errorf("store worker unreachable: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return CrawlResult{}, fmt.Errorf("store crawl %d: %s", resp.StatusCode, string(body))
	}
	var result CrawlResult
	if err := json.Unmarshal(body, &result); err != nil {
		return CrawlResult{}, err
	}
	return result, nil
}

func (c *Client) CrawlApps(ctx context.Context, appName, playAppID, appStoreAppID string, limit int) (CrawlResult, error) {
	return c.Crawl(ctx, crawlRequest{
		AppName:       appName,
		PlayAppID:     playAppID,
		AppStoreAppID: appStoreAppID,
		Limit:         limit,
		Lang:          "tr",
		Country:       "tr",
	})
}
