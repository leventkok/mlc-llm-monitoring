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
	mux.HandleFunc("/v1/chat/completions", withAPIKey(handleChat))

	log.Printf("mlc-mock listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func withAPIKey(next http.HandlerFunc) http.HandlerFunc {
	expected := strings.TrimSpace(os.Getenv("MLC_API_KEY"))
	return func(w http.ResponseWriter, r *http.Request) {
		if expected != "" && r.Header.Get("X-MLC-API-Key") != expected {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
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
	sentiment = detectSentiment(text)
	category = detectCategory(text, sentiment)
	if sentiment == "neutral" && (category == "bug" || category == "spam") {
		sentiment = "negative"
	}
	return category, sentiment
}

func detectSentiment(text string) string {
	for _, w := range []string{
		"crash", "bug", "broken", "bad", "terrible", "awful", "hate", "worst", "horrible",
		"disappoint", "sucks", "poor", "useless", "slow", "freeze", "error", "fail", "garbage",
		"kötü", "berbat", "rezalet", "iğrenç", "beğenmedim", "tavsiye etmem", "çöp",
		"sinir", "hayal kırıklığı", "bok", "saçma", "korkunç", "felaket",
	} {
		if strings.Contains(text, w) {
			return "negative"
		}
	}
	for _, w := range []string{
		"love", "great", "awesome", "excellent", "amazing", "perfect", "best", "fantastic",
		"güzel", "harika", "mükemmel", "seviyorum", "beğendim", "süper", "muhteşem", "bayıldım",
	} {
		if strings.Contains(text, w) {
			return "positive"
		}
	}
	return "neutral"
}

func detectCategory(text, sentiment string) string {
	for _, w := range []string{
		"crash", "bug", "broken", "error", "fail", "freeze", "not working", "doesn't work",
		"çök", "hata", "çalışmıyor", "donuyor", "yavaş", "açılmıyor",
	} {
		if strings.Contains(text, w) {
			return "bug"
		}
	}
	for _, w := range []string{
		"feature", "request", "please add", "would like", "wish", "should add",
		"özellik", "ekle", "istiyorum", "olsun", "eklenmeli",
	} {
		if strings.Contains(text, w) {
			return "feature"
		}
	}
	for _, w := range []string{
		"love", "great", "awesome", "excellent", "amazing", "perfect", "best", "fantastic",
		"güzel", "harika", "mükemmel", "seviyorum", "beğendim", "süper", "muhteşem", "bayıldım",
	} {
		if strings.Contains(text, w) {
			return "praise"
		}
	}
	for _, w := range []string{"spam", "fake", "scam", "click here", "free money", "sahte"} {
		if strings.Contains(text, w) {
			return "spam"
		}
	}
	if sentiment == "positive" {
		return "praise"
	}
	return "other"
}
