package rdap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	b := &Bootstrap{servers: map[string]string{"com": server.URL}}
	client := NewClient(server.Client(), b, 0)

	result := client.Check(context.Background(), "available.com")
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if !result.Available {
		t.Error("expected domain to be available")
	}
	if result.Domain != "available.com" {
		t.Errorf("domain = %q, want available.com", result.Domain)
	}
}

func TestCheckRegistered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rdap+json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"objectClassName":"domain","ldhName":"google.com"}`))
	}))
	defer server.Close()

	b := &Bootstrap{servers: map[string]string{"com": server.URL}}
	client := NewClient(server.Client(), b, 0)

	result := client.Check(context.Background(), "google.com")
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Available {
		t.Error("expected domain to be registered (not available)")
	}
}

func TestCheckRateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	b := &Bootstrap{servers: map[string]string{"com": server.URL}}
	client := NewClient(server.Client(), b, 2)

	result := client.Check(context.Background(), "test.com")
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if !result.Available {
		t.Error("expected domain to be available after retries")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestCheckUnknownTLD(t *testing.T) {
	b := &Bootstrap{servers: map[string]string{}}
	httpClient := &http.Client{Timeout: 5 * time.Second}
	client := NewClient(httpClient, b, 0)

	result := client.Check(context.Background(), "test.unknowntld")
	if result.Error == nil {
		t.Error("expected error for unknown TLD")
	}
}

func TestCheckContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer server.Close()

	b := &Bootstrap{servers: map[string]string{"com": server.URL}}
	client := NewClient(server.Client(), b, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result := client.Check(ctx, "slow.com")
	if result.Error == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestExtractTLD(t *testing.T) {
	tests := []struct {
		domain string
		want   string
	}{
		{"example.com", "com"},
		{"test.co.uk", "uk"},
		{"hello.ai", "ai"},
	}
	for _, tt := range tests {
		got := extractTLD(tt.domain)
		if got != tt.want {
			t.Errorf("extractTLD(%q) = %q, want %q", tt.domain, got, tt.want)
		}
	}
}
