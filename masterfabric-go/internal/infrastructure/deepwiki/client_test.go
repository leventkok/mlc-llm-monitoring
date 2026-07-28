package deepwiki_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leventkok/mlc-llm-monitoring/masterfabric-go/internal/infrastructure/deepwiki"
)

func TestClient_AskQuestion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`event: message
data: {"jsonrpc":"2.0","id":1,"result":{"structuredContent":{"result":"hello from wiki"}}}

`))
	}))
	defer srv.Close()

	client := deepwiki.NewClient(srv.URL)
	text, err := client.AskQuestion(context.Background(), "mlc-ai/mlc-llm", "test?")
	if err != nil {
		t.Fatal(err)
	}
	if text != "hello from wiki" {
		t.Fatalf("got %q", text)
	}
}
