package pfsense_rest_v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetFirewallRules_HTTPTest_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/v2/firewall/rules" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "0" {
			t.Fatalf("unexpected limit query value: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":200,"status":"ok","data":[{"type":"pass","interface":["wan"],"disabled":false,"ipprotocol":"inet","log":true,"descr":"allow web","protocol":"tcp","source":"any","source_port":"any","destination":"10.0.0.2","destination_port":"443"}]}`))
	}))
	defer server.Close()

	client, err := NewPFSenseClientV2(server.URL, &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	rules, err := client.GetFirewallRules(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Description != "allow web" {
		t.Fatalf("unexpected rule description: got %q", rules[0].Description)
	}
	if rules[0].Destination != "10.0.0.2" {
		t.Fatalf("unexpected destination: got %q", rules[0].Destination)
	}
}

func TestGetFirewallRules_HTTPTest_UnhappyPath(t *testing.T) {
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

	_, err = client.GetFirewallRules(context.Background())
	if err == nil {
		t.Fatal("expected error from non-200 response")
	}
	if !strings.Contains(err.Error(), "unexpected response retrieving firewall rules") {
		t.Fatalf("unexpected error: %v", err)
	}
}
