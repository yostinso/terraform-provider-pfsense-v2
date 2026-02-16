package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPFSenseDataSource_FirewallRulesApply(t *testing.T) {
	const apiToken = "test-token"

	var (
		mu                    sync.RWMutex
		firewallRulesResponse = `{"code":200,"data":[{"type":"pass","interface":["lan"],"disabled":false,"ipprotocol":"inet","log":true,"descr":"allow web","protocol":"tcp","source":"any","source_port":"1024:65535","destination":"lan","destination_port":"443"}]}`
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != apiToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":401,"message":"unauthorized"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/system/hostname":
			_, _ = w.Write([]byte(`{"code":200,"data":{"hostname":"ng4100","domain":"example.local"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/firewall/rules":
			mu.RLock()
			payload := firewallRulesResponse
			mu.RUnlock()
			_, _ = w.Write([]byte(payload))
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
				Config: testAccPFSenseFirewallRulesConfig(server.URL, apiToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "hostname", "ng4100"),
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "firewall_rules.#", "1"),
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "firewall_rules.0.type", "pass"),
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "firewall_rules.0.interfaces.#", "1"),
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "firewall_rules.0.interfaces.0", "lan"),
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "firewall_rules.0.description", "allow web"),
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "firewall_rules.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "firewall_rules.0.destination_port", "443"),
				),
			},
			{
				PreConfig: func() {
					mu.Lock()
					firewallRulesResponse = `{"code":200,"data":[]}`
					mu.Unlock()
				},
				Config: testAccPFSenseFirewallRulesConfig(server.URL, apiToken),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "hostname", "ng4100"),
					resource.TestCheckResourceAttr("data.scaffolding_configs.test", "firewall_rules.#", "0"),
				),
			},
		},
	})
}

func testAccPFSenseFirewallRulesConfig(url string, token string) string {
	return fmt.Sprintf(`
provider "scaffolding" {
  url              = %q
  insecure         = false
  api_client_token = %q
}

data "scaffolding_configs" "test" {}
`, url, token)
}
