package pfsense_rest_v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetBaseConfig_HTTPTest_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/v2/system/hostname" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":200,"status":"ok","data":{"hostname":"pfsense","domain":"example.local"}}`))
	}))
	defer server.Close()

	client, err := NewPFSenseClientV2(server.URL, &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	cfg, err := client.GetBaseConfig(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Hostname != "pfsense" {
		t.Fatalf("unexpected hostname: got %q", cfg.Hostname)
	}
	if cfg.Domain != "example.local" {
		t.Fatalf("unexpected domain: got %q", cfg.Domain)
	}
}

func TestGetBaseConfig_HTTPTest_UnhappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":400,"status":"bad request","message":"invalid"}`))
	}))
	defer server.Close()

	client, err := NewPFSenseClientV2(server.URL, &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.GetBaseConfig(context.Background())
	if err == nil {
		t.Fatal("expected error from non-200 response")
	}
	if !strings.Contains(err.Error(), "unexpected response retrieving base config") {
		t.Fatalf("unexpected error: %v", err)
	}
}
