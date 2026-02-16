package pfsense_rest_v2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func parseFirewallRulesResponse(t *testing.T, statusCode int, body string) *GetFirewallRulesEndpointResponse {
	t.Helper()

	parsed, err := ParseGetFirewallRulesEndpointResponse(&http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	})
	if err != nil {
		t.Fatalf("failed to parse test response: %v", err)
	}

	return parsed
}

func TestGetFirewallRules_Unit(t *testing.T) {
	successResponse := parseFirewallRulesResponse(
		t,
		http.StatusOK,
		`{"code":200,"status":"ok","data":[{"type":"pass","interface":["wan"],"disabled":false,"ipprotocol":"inet","log":true,"descr":"allow web","protocol":"tcp","source":"any","source_port":"any","destination":"10.0.0.2","destination_port":"443"}]}`,
	)
	emptyResponse := parseFirewallRulesResponse(
		t,
		http.StatusOK,
		`{"code":200,"status":"ok","data":[]}`,
	)
	unexpectedResponse := parseFirewallRulesResponse(
		t,
		http.StatusBadRequest,
		`{"code":400,"status":"bad request","message":"invalid"}`,
	)

	testCases := []struct {
		name      string
		response  *GetFirewallRulesEndpointResponse
		apiErr    error
		wantCount int
		wantError string
	}{
		{
			name:      "maps rule fields on success",
			response:  successResponse,
			wantCount: 1,
		},
		{
			name:      "returns empty list for empty data",
			response:  emptyResponse,
			wantCount: 0,
		},
		{
			name:      "returns error when api client errors",
			apiErr:    errors.New("boom"),
			wantError: "boom",
		},
		{
			name:      "returns error on nil response",
			response:  nil,
			wantError: "unexpected nil response retrieving firewall rules",
		},
		{
			name:      "returns error on unexpected response",
			response:  unexpectedResponse,
			wantError: "unexpected response retrieving firewall rules",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &PFSenseClientV2{
				apiClient: &fakePFSenseAPIClient{
					getFirewallRulesEndpointWithResponse: func(ctx context.Context, params *GetFirewallRulesEndpointParams, reqEditors ...RequestEditorFn) (*GetFirewallRulesEndpointResponse, error) {
						if params == nil || params.Limit == nil || *params.Limit != 0 {
							t.Fatalf("expected limit query param of 0, got: %#v", params)
						}
						if tc.apiErr != nil {
							return nil, tc.apiErr
						}
						return tc.response, nil
					},
				},
			}

			rules, err := client.GetFirewallRules(context.Background())

			if tc.wantError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantError)
				}
				if !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected error containing %q, got %q", tc.wantError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if len(rules) != tc.wantCount {
				t.Fatalf("unexpected number of rules: got %d want %d", len(rules), tc.wantCount)
			}

			if tc.wantCount == 1 {
				rule := rules[0]
				if rule.Type == nil || *rule.Type != "pass" {
					t.Fatalf("unexpected rule type: %#v", rule.Type)
				}
				if len(rule.Interfaces) != 1 || rule.Interfaces[0] != "wan" {
					t.Fatalf("unexpected interfaces: %#v", rule.Interfaces)
				}
				if rule.Protocol == nil || *rule.Protocol != "tcp" {
					t.Fatalf("unexpected protocol: %#v", rule.Protocol)
				}
				if rule.AddressFamily != "inet" {
					t.Fatalf("unexpected address family: %q", rule.AddressFamily)
				}
				if rule.Description != "allow web" {
					t.Fatalf("unexpected description: %q", rule.Description)
				}
				if rule.Source != "any" || rule.Destination != "10.0.0.2" {
					t.Fatalf("unexpected source/destination: %q -> %q", rule.Source, rule.Destination)
				}
				if rule.SourcePort == nil || *rule.SourcePort != "any" {
					t.Fatalf("unexpected source port: %#v", rule.SourcePort)
				}
				if rule.DestinationPort == nil || *rule.DestinationPort != "443" {
					t.Fatalf("unexpected destination port: %#v", rule.DestinationPort)
				}
			}
		})
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
