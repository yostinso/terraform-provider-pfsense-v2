package pfsense_rest_v2

import (
	"context"
	"testing"
)

func TestGetBaseConfig_NoPanicOnClientError(t *testing.T) {
	client, err := NewPFSenseClientV2("http://127.0.0.1:1", &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.GetBaseConfig(context.Background())
	if err == nil {
		t.Fatal("expected error from unreachable client")
	}
}

func TestGetFirewallRules_NoPanicOnClientError(t *testing.T) {
	client, err := NewPFSenseClientV2("http://127.0.0.1:1", &APIKeyAuth{APIToken: "test"}, false)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.GetFirewallRules(context.Background())
	if err == nil {
		t.Fatal("expected error from unreachable client")
	}
}
