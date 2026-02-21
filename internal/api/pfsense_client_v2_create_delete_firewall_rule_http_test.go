package pfsense_rest_v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateFirewallRule_HTTPTest_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/v2/firewall/rule" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["type"] != "pass" {
			t.Fatalf("unexpected type in request body: %v", body["type"])
		}
		if body["destination"] != "10.0.0.1" {
			t.Fatalf("unexpected destination in request body: %v", body["destination"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":200,"status":"ok","data":{"id":7,"type":"pass","interface":["lan"],"disabled":false,"ipprotocol":"inet","log":true,"descr":"allow https","protocol":"tcp","source":"any","source_port":null,"destination":"10.0.0.1","destination_port":"443"}}`))
	}))
	defer server.Close()

	client, err := NewPFSenseClientV2(server.URL, &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	proto := "tcp"
	dstPort := "443"
	rule, err := client.CreateFirewallRule(context.Background(), &PFSenseFirewallRule{
		Type:            strPtr("pass"),
		Interfaces:      []string{"lan"},
		AddressFamily:   "inet",
		Source:          "any",
		Destination:     "10.0.0.1",
		Description:     "allow https",
		Protocol:        &proto,
		DestinationPort: &dstPort,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rule.ID == nil || *rule.ID != 7 {
		t.Fatalf("unexpected ID: %#v", rule.ID)
	}
	if rule.Type == nil || *rule.Type != "pass" {
		t.Fatalf("unexpected type: %#v", rule.Type)
	}
	if rule.Description != "allow https" {
		t.Fatalf("unexpected description: %q", rule.Description)
	}
	if rule.DestinationPort == nil || *rule.DestinationPort != "443" {
		t.Fatalf("unexpected destination port: %#v", rule.DestinationPort)
	}
}

func TestCreateFirewallRule_HTTPTest_UnhappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":400,"status":"bad request","message":"invalid input"}`))
	}))
	defer server.Close()

	client, err := NewPFSenseClientV2(server.URL, &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.CreateFirewallRule(context.Background(), &PFSenseFirewallRule{
		Type:          strPtr("pass"),
		Interfaces:    []string{"lan"},
		AddressFamily: "inet",
		Source:        "any",
		Destination:   "any",
	})
	if err == nil {
		t.Fatal("expected error from non-200 response")
	}
	if !strings.Contains(err.Error(), "unexpected response creating firewall rule") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteFirewallRule_HTTPTest_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/v2/firewall/rule" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("id"); got != "42" {
			t.Fatalf("unexpected id query param: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":200,"status":"ok","data":{"id":42,"type":"pass","interface":["lan"],"disabled":false,"ipprotocol":"inet","log":true,"descr":"allow https","protocol":"tcp","source":"any","source_port":null,"destination":"10.0.0.1","destination_port":"443"}}`))
	}))
	defer server.Close()

	client, err := NewPFSenseClientV2(server.URL, &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.DeleteFirewallRule(context.Background(), 42); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestDeleteFirewallRule_HTTPTest_UnhappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":400,"status":"bad request","message":"rule not found"}`))
	}))
	defer server.Close()

	client, err := NewPFSenseClientV2(server.URL, &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.DeleteFirewallRule(context.Background(), 42)
	if err == nil {
		t.Fatal("expected error from non-200 response")
	}
	if !strings.Contains(err.Error(), "unexpected response deleting firewall rule") {
		t.Fatalf("unexpected error: %v", err)
	}
}
