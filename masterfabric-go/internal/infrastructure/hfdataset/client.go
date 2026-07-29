package hfdataset

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	datasetModel "github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/domain/dataset/model"
)

const defaultServer = "https://datasets-server.huggingface.co"

// Client reads InferReview review rows from Hugging Face Datasets Server or dataset-worker CSV.
type Client struct {
	datasetID  string
	token      string
	serverURL  string
	workerURL  string
	httpClient *http.Client
}

func NewClient(datasetID, token, serverURL, workerURL string) *Client {
	datasetID = strings.TrimSpace(datasetID)
	if datasetID == "" {
		datasetID = "levonov/inferreview-app-reviews"
	}
	serverURL = strings.TrimRight(strings.TrimSpace(serverURL), "/")
	if serverURL == "" {
		serverURL = defaultServer
	}
	return &Client{
		datasetID: datasetID,
		token:     strings.TrimSpace(token),
		serverURL: serverURL,
		workerURL: strings.TrimRight(strings.TrimSpace(workerURL), "/"),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) DatasetID() string {
	return c.datasetID
}

type firstRowsResponse struct {
	Rows []struct {
		Row map[string]any `json:"row"`
	} `json:"rows"`
}

func (c *Client) ListRows(ctx context.Context, offset, limit int) (datasetModel.ReviewPage, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	need := offset + limit
	if need > 500 {
		need = 500
	}

	rawRows, err := c.fetchFirstRows(ctx, need)
	if err != nil && c.workerURL != "" {
		rawRows, err = c.fetchWorkerJSON(ctx, offset, limit)
		if err != nil {
			return datasetModel.ReviewPage{}, err
		}
		return datasetModel.ReviewPage{
			DatasetID: c.datasetID,
			Offset:    offset,
			Limit:     limit,
			Rows:      rawRows,
		}, nil
	}
	if err != nil {
		return datasetModel.ReviewPage{}, err
	}

	if offset > len(rawRows) {
		rawRows = nil
	} else {
		end := offset + limit
		if end > len(rawRows) {
			end = len(rawRows)
		}
		rawRows = rawRows[offset:end]
	}

	return datasetModel.ReviewPage{
		DatasetID: c.datasetID,
		Offset:    offset,
		Limit:     limit,
		Rows:      rawRows,
	}, nil
}

func (c *Client) ExportCSV(ctx context.Context, offset, limit int) ([]byte, error) {
	if c.workerURL != "" {
		u := fmt.Sprintf("%s/export.csv?offset=%d&limit=%d", c.workerURL, offset, limit)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			b, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("dataset worker %d: %s", resp.StatusCode, string(b))
		}
		return io.ReadAll(resp.Body)
	}

	page, err := c.ListRows(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	return encodeCSV(page.Rows)
}

func (c *Client) fetchFirstRows(ctx context.Context, length int) ([]datasetModel.ReviewRow, error) {
	q := url.Values{}
	q.Set("dataset", c.datasetID)
	q.Set("config", "default")
	q.Set("split", "train")
	q.Set("length", strconv.Itoa(length))

	reqURL := c.serverURL + "/first-rows?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hf datasets request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hf datasets %d: %s", resp.StatusCode, string(body))
	}

	var parsed firstRowsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode hf response: %w", err)
	}

	out := make([]datasetModel.ReviewRow, 0, len(parsed.Rows))
	for _, item := range parsed.Rows {
		out = append(out, mapRow(item.Row))
	}
	return out, nil
}

func (c *Client) fetchWorkerJSON(ctx context.Context, offset, limit int) ([]datasetModel.ReviewRow, error) {
	u := fmt.Sprintf("%s/rows?offset=%d&limit=%d", c.workerURL, offset, limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("dataset worker %d: %s", resp.StatusCode, string(body))
	}
	var payload struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	out := make([]datasetModel.ReviewRow, 0, len(payload.Rows))
	for _, row := range payload.Rows {
		out = append(out, mapRow(row))
	}
	return out, nil
}

func mapRow(row map[string]any) datasetModel.ReviewRow {
	rating := 0
	switch v := row["rating"].(type) {
	case float64:
		rating = int(v)
	case string:
		rating, _ = strconv.Atoi(v)
	}
	return datasetModel.ReviewRow{
		ReviewID:   asString(row["review_id"]),
		Store:      asString(row["store"]),
		AppName:    asString(row["app_name"]),
		AppVersion: asString(row["app_version"]),
		Rating:     rating,
		Text:       asString(row["text"]),
		Language:   asString(row["language"]),
		Category:   asString(row["category"]),
		Sentiment:  asString(row["sentiment"]),
	}
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func encodeCSV(rows []datasetModel.ReviewRow) ([]byte, error) {
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{
		"review_id", "store", "app_name", "app_version", "rating",
		"text", "language", "category", "sentiment",
	})
	for _, r := range rows {
		_ = w.Write([]string{
			r.ReviewID, r.Store, r.AppName, r.AppVersion, strconv.Itoa(r.Rating),
			r.Text, r.Language, r.Category, r.Sentiment,
		})
	}
	w.Flush()
	return []byte(b.String()), w.Error()
}
