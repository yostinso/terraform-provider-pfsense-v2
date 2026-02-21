package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPFSenseFirewallRuleResource_CreateDelete(t *testing.T) {
	const apiToken = "test-token"
	const ruleID = 99

	// storedRule holds the server-side rule after POST; nil means it does not exist.
	var (
		mu         sync.RWMutex
		storedRule map[string]interface{}
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != apiToken {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":401,"message":"unauthorized"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/system/hostname":
			_, _ = w.Write([]byte(`{"code":200,"data":{"hostname":"ng4100","domain":"example.local"}}`))

		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/firewall/rule":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"code":400,"message":"bad request"}`))
				return
			}
			mu.Lock()
			storedRule = map[string]interface{}{
				"id":               ruleID,
				"type":             body["type"],
				"interface":        body["interface"],
				"disabled":         false,
				"ipprotocol":       body["ipprotocol"],
				"log":              false,
				"descr":            body["descr"],
				"protocol":         body["protocol"],
				"source":           body["source"],
				"source_port":      nil,
				"destination":      body["destination"],
				"destination_port": body["destination_port"],
			}
			rule := storedRule
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 200, "status": "ok", "data": rule})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/firewall/rule":
			mu.RLock()
			rule := storedRule
			mu.RUnlock()
			if rule == nil || r.URL.Query().Get("id") != fmt.Sprintf("%d", ruleID) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"code":404,"message":"not found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 200, "status": "ok", "data": rule})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/firewall/rule":
			mu.Lock()
			rule := storedRule
			storedRule = nil
			mu.Unlock()
			if rule == nil {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"code":404,"message":"not found"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 200, "status": "ok", "data": rule})

		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":404,"message":"not found"}`))
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: apply creates the rule; the framework destroys it at teardown,
				// exercising the DELETE path automatically.
				Config: testAccFirewallRuleResourceConfig(server.URL, apiToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("pfsense-v2_firewall_rule.test", "id", fmt.Sprintf("%d", ruleID)),
					resource.TestCheckResourceAttr("pfsense-v2_firewall_rule.test", "type", "pass"),
					resource.TestCheckResourceAttr("pfsense-v2_firewall_rule.test", "interfaces.0", "lan"),
					resource.TestCheckResourceAttr("pfsense-v2_firewall_rule.test", "description", "test rule"),
					resource.TestCheckResourceAttr("pfsense-v2_firewall_rule.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("pfsense-v2_firewall_rule.test", "destination_port", "443"),
				),
			},
		},
	})
}

func testAccFirewallRuleResourceConfig(url string, token string) string {
	return fmt.Sprintf(`
provider "scaffolding" {
  url              = %q
  insecure         = false
  api_client_token = %q
}

resource "pfsense-v2_firewall_rule" "test" {
  type             = "pass"
  interfaces       = ["lan"]
  address_family   = "inet"
  source           = "any"
  destination      = "any"
  description      = "test rule"
  protocol         = "tcp"
  destination_port = "443"
}
`, url, token)
}
