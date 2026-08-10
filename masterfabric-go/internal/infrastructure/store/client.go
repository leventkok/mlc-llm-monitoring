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

// Warmup pings the worker (Render cold start). Errors are ignored.
func (c *Client) Warmup(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func sanitizeStoreErr(action string, status int, body []byte) error {
	msg := strings.TrimSpace(string(body))
	if strings.Contains(strings.ToLower(msg), "<!doctype") || strings.Contains(strings.ToLower(msg), "<html") {
		if status == 502 || status == 503 || status == 504 {
			return fmt.Errorf("store worker geçici olarak yanıt vermiyor (Render cold start). 1–2 dakika bekleyip tekrar deneyin")
		}
		return fmt.Errorf("store worker hatası (%d)", status)
	}
	if len(msg) > 240 {
		msg = msg[:240] + "…"
	}
	if msg == "" {
		return fmt.Errorf("store %s failed with status %d", action, status)
	}
	return fmt.Errorf("store %s %d: %s", action, status, msg)
}

func retryableStatus(status int) bool {
	return status == 502 || status == 503 || status == 504
}

func (c *Client) doGET(ctx context.Context, rawURL string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("store worker unreachable: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

func (c *Client) doPOST(ctx context.Context, path string, payload []byte) ([]byte, int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("store worker unreachable: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
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
	target := u.String()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * 8 * time.Second):
			}
		}
		body, status, err := c.doGET(ctx, target)
		if err != nil {
			lastErr = err
			continue
		}
		if status >= 400 {
			lastErr = sanitizeStoreErr("search", status, body)
			if retryableStatus(status) {
				continue
			}
			return nil, lastErr
		}
		var parsed struct {
			Apps []AppResult `json:"apps"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}
		return parsed.Apps, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("store search failed after retries")
}

func (c *Client) Crawl(ctx context.Context, req crawlRequest) (CrawlResult, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return CrawlResult{}, err
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return CrawlResult{}, ctx.Err()
			case <-time.After(time.Duration(attempt) * 8 * time.Second):
			}
		}
		body, status, err := c.doPOST(ctx, "/crawl", payload)
		if err != nil {
			lastErr = err
			continue
		}
		if status >= 400 {
			lastErr = sanitizeStoreErr("crawl", status, body)
			if retryableStatus(status) {
				continue
			}
			return CrawlResult{}, lastErr
		}
		var result CrawlResult
		if err := json.Unmarshal(body, &result); err != nil {
			return CrawlResult{}, err
		}
		return result, nil
	}
	if lastErr != nil {
		return CrawlResult{}, lastErr
	}
	return CrawlResult{}, fmt.Errorf("store crawl failed after retries")
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
	// `limit` is a legacy/fallback cap (max 10000 in store-worker schema), not the sum of per-store limits.
	combinedLimit := 0
	if opts.PlayAppID != "" {
		combinedLimit += playLimit
	}
	if opts.AppStoreAppID != "" {
		combinedLimit += appStoreLimit
	}
	if combinedLimit <= 0 {
		combinedLimit = 500
	}
	fallbackLimit := playLimit
	if appStoreLimit > fallbackLimit {
		fallbackLimit = appStoreLimit
	}
	if fallbackLimit <= 0 {
		if combinedLimit > 10000 {
			fallbackLimit = 10000
		} else {
			fallbackLimit = combinedLimit
		}
	}
	if fallbackLimit > 10000 {
		fallbackLimit = 10000
	}

	req := crawlRequest{
		AppName:       opts.AppName,
		PlayAppID:     opts.PlayAppID,
		AppStoreAppID: opts.AppStoreAppID,
		Limit:         fallbackLimit,
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
