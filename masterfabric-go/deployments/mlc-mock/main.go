package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type chatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func main() {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/v1/chat/completions", handleChat)

	log.Printf("mlc-mock listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	reviewText := extractReviewText(req.Messages)
	category, sentiment := classify(reviewText)
	payload, _ := json.Marshal(map[string]string{
		"category":  category,
		"sentiment": sentiment,
	})

	// Simulate inference latency for local load tests.
	time.Sleep(150 * time.Millisecond)

	resp := chatResponse{}
	resp.Choices = []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}{
		{Message: struct {
			Content string `json:"content"`
		}{Content: string(payload)}},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func extractReviewText(messages []struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}) string {
	for _, m := range messages {
		const marker = "Review:"
		idx := strings.Index(m.Content, marker)
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(m.Content[idx+len(marker):])
		rest = strings.TrimPrefix(rest, `"`)
		if end := strings.LastIndex(rest, `"`); end >= 0 {
			rest = rest[:end]
		}
		return strings.ToLower(rest)
	}
	return ""
}

func classify(text string) (category, sentiment string) {
	category = "other"
	sentiment = "neutral"

	switch {
	case strings.Contains(text, "crash") || strings.Contains(text, "bug") || strings.Contains(text, "broken"):
		return "bug", "negative"
	case strings.Contains(text, "feature") || strings.Contains(text, "request") || strings.Contains(text, "add"):
		return "feature", "neutral"
	case strings.Contains(text, "love") || strings.Contains(text, "great") || strings.Contains(text, "awesome"):
		return "praise", "positive"
	case strings.Contains(text, "spam") || strings.Contains(text, "fake"):
		return "spam", "negative"
	}
	return category, sentiment
}
