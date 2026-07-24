package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"
)

const defaultAPIURL = "https://mlc-llm-monitoring.onrender.com"

type apiClient struct {
	baseURL    string
	httpClient *http.Client
	bearer     string
	mu         sync.RWMutex
}

func newAPIClient() (*apiClient, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("INFERREVIEW_API_URL")), "/")
	if baseURL == "" {
		baseURL = defaultAPIURL
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("cookie jar: %w", err)
	}

	c := &apiClient{
		baseURL: baseURL,
		bearer:  strings.TrimSpace(os.Getenv("INFERREVIEW_JWT_TOKEN")),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
			Jar:     jar,
		},
	}

	if email := strings.TrimSpace(os.Getenv("INFERREVIEW_EMAIL")); email != "" {
		password := os.Getenv("INFERREVIEW_PASSWORD")
		if password == "" {
			return nil, fmt.Errorf("INFERREVIEW_PASSWORD is required when INFERREVIEW_EMAIL is set")
		}
		if err := c.login(email, password); err != nil {
			return nil, fmt.Errorf("startup login failed: %w", err)
		}
	}

	return c, nil
}

func (c *apiClient) login(email, password string) error {
	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return err
	}

	resp, err := c.do(http.MethodPost, "/auth/login", body, false)
	if err != nil {
		return err
	}

	var msg struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	_ = json.Unmarshal(resp, &msg)
	if msg.Error != "" {
		return fmt.Errorf("%s", msg.Error)
	}
	return nil
}

func (c *apiClient) setBearer(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bearer = strings.TrimSpace(token)
}

func (c *apiClient) get(path string) ([]byte, error) {
	return c.do(http.MethodGet, path, nil, true)
}

func (c *apiClient) post(path string, payload any) ([]byte, error) {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	return c.do(http.MethodPost, path, body, true)
}

func (c *apiClient) do(method, path string, body []byte, auth bool) ([]byte, error) {
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, u, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if auth {
		c.mu.RLock()
		token := c.bearer
		c.mu.RUnlock()
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errBody struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(data, &errBody)
		if errBody.Error != "" {
			return nil, fmt.Errorf("API %d: %s", resp.StatusCode, errBody.Error)
		}
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return data, nil
}

func prettyJSON(raw []byte) (string, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw), nil
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return string(raw), nil
	}
	return string(out), nil
}
