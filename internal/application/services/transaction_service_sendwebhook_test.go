package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendWebhook_EmptyURL(t *testing.T) {
	svc := &TransactionService{}
	// Should return immediately without panic
	svc.sendWebhook("", "tron", map[string]interface{}{"status": "pending"})
}

func TestSendWebhook_MarshalError(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer server.Close()
	svc := &TransactionService{}
	// Include a function in map to force json.Marshal error
	svc.sendWebhook(server.URL, "tron", map[string]interface{}{"bad": func() {}})
	if called {
		t.Fatalf("expected no HTTP call due to marshal error")
	}
}
