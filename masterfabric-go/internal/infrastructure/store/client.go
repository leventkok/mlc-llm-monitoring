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
			Timeout: 600 * time.Second,
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

type CrawlOptions struct {
	AppName             string
	PlayAppID           string
	AppStoreAppID       string
	PlayReviewLimit     int
	AppStoreReviewLimit int
	Lang                string
	Country             string
}

type crawlRequest struct {
	AppName         string `json:"app_name"`
	PlayAppID       string `json:"play_app_id"`
	AppStoreAppID   string `json:"appstore_app_id"`
	Limit           int    `json:"limit"`
	PlayLimit       *int   `json:"play_limit,omitempty"`
	AppStoreLimit   *int   `json:"appstore_limit,omitempty"`
	Lang            string `json:"lang"`
	Country         string `json:"country"`
}

func (c *Client) Search(ctx context.Context, store, query, country, lang string, limit int) ([]AppResult, error) {
	u, err := url.Parse(c.baseURL + "/search")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("store", store)
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	if strings.TrimSpace(country) != "" {
		q.Set("country", strings.TrimSpace(country))
	}
	if strings.TrimSpace(lang) != "" {
		q.Set("lang", strings.TrimSpace(lang))
	}
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

func (c *Client) CrawlApps(ctx context.Context, opts CrawlOptions) (CrawlResult, error) {
	lang := strings.TrimSpace(opts.Lang)
	if lang == "" {
		lang = "tr"
	}
	country := strings.TrimSpace(opts.Country)
	if country == "" {
		country = "tr"
	}

	playLimit := opts.PlayReviewLimit
	appStoreLimit := opts.AppStoreReviewLimit
	totalLimit := 0
	if opts.PlayAppID != "" {
		totalLimit += playLimit
	}
	if opts.AppStoreAppID != "" {
		totalLimit += appStoreLimit
	}
	if totalLimit <= 0 {
		totalLimit = 500
	}

	req := crawlRequest{
		AppName:       opts.AppName,
		PlayAppID:     opts.PlayAppID,
		AppStoreAppID: opts.AppStoreAppID,
		Limit:         totalLimit,
		Lang:          lang,
		Country:       country,
	}
	if opts.PlayAppID != "" && playLimit > 0 {
		pl := playLimit
		req.PlayLimit = &pl
	}
	if opts.AppStoreAppID != "" && appStoreLimit > 0 {
		al := appStoreLimit
		req.AppStoreLimit = &al
	}
	return c.Crawl(ctx, req)
}
