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

// representative 200 JSON returned by POST /api/v2/firewall/rule
const createFirewallRuleSuccessJSON = `{"code":200,"status":"ok","data":{"id":42,"type":"pass","interface":["lan"],"disabled":false,"ipprotocol":"inet","log":true,"descr":"allow https","protocol":"tcp","source":"any","source_port":null,"destination":"10.0.0.1","destination_port":"443"}}`

// representative 200 JSON returned by DELETE /api/v2/firewall/rule
const deleteFirewallRuleSuccessJSON = `{"code":200,"status":"ok","data":{"id":42,"type":"pass","interface":["lan"],"disabled":false,"ipprotocol":"inet","log":true,"descr":"allow https","protocol":"tcp","source":"any","source_port":null,"destination":"10.0.0.1","destination_port":"443"}}`

func parsePostFirewallRuleResponse(t *testing.T, statusCode int, body string) *PostFirewallRuleEndpointResponse {
	t.Helper()
	parsed, err := ParsePostFirewallRuleEndpointResponse(&http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	})
	if err != nil {
		t.Fatalf("failed to parse test response: %v", err)
	}
	return parsed
}

func parseDeleteFirewallRuleResponse(t *testing.T, statusCode int, body string) *DeleteFirewallRuleEndpointResponse {
	t.Helper()
	parsed, err := ParseDeleteFirewallRuleEndpointResponse(&http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	})
	if err != nil {
		t.Fatalf("failed to parse test response: %v", err)
	}
	return parsed
}

func TestCreateFirewallRule_Unit(t *testing.T) {
	successResponse := parsePostFirewallRuleResponse(t, http.StatusOK, createFirewallRuleSuccessJSON)
	unexpectedResponse := parsePostFirewallRuleResponse(t, http.StatusBadRequest,
		`{"code":400,"status":"bad request","message":"invalid input"}`)

	proto := "tcp"
	dstPort := "443"
	input := &PFSenseFirewallRule{
		Type:            strPtr("pass"),
		Interfaces:      []string{"lan"},
		AddressFamily:   "inet",
		Source:          "any",
		Destination:     "10.0.0.1",
		Description:     "allow https",
		Protocol:        &proto,
		DestinationPort: &dstPort,
		Disabled:        false,
		Log:             true,
	}

	testCases := []struct {
		name      string
		response  *PostFirewallRuleEndpointResponse
		apiErr    error
		wantError string
	}{
		{
			name:     "maps created rule fields on success",
			response: successResponse,
		},
		{
			name:      "returns error when api client errors",
			apiErr:    errors.New("network failure"),
			wantError: "network failure",
		},
		{
			name:      "returns error on nil response",
			response:  nil,
			wantError: "unexpected nil response creating firewall rule",
		},
		{
			name:      "returns error on unexpected response",
			response:  unexpectedResponse,
			wantError: "unexpected response creating firewall rule",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var capturedBody PostFirewallRuleEndpointJSONRequestBody
			client := &PFSenseClientV2{
				apiClient: &fakePFSenseAPIClient{
					postFirewallRuleEndpointWithResponse: func(ctx context.Context, body PostFirewallRuleEndpointJSONRequestBody, reqEditors ...RequestEditorFn) (*PostFirewallRuleEndpointResponse, error) {
						capturedBody = body
						if tc.apiErr != nil {
							return nil, tc.apiErr
						}
						return tc.response, nil
					},
				},
			}

			got, err := client.CreateFirewallRule(context.Background(), input)

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

			// Verify request fields were mapped correctly.
			if capturedBody.Type != PostFirewallRuleEndpointJSONBodyTypePass {
				t.Fatalf("unexpected type in request: %q", capturedBody.Type)
			}
			if len(capturedBody.Interface) != 1 || capturedBody.Interface[0] != "lan" {
				t.Fatalf("unexpected interface in request: %v", capturedBody.Interface)
			}
			if capturedBody.Source != "any" {
				t.Fatalf("unexpected source in request: %q", capturedBody.Source)
			}
			if capturedBody.Destination != "10.0.0.1" {
				t.Fatalf("unexpected destination in request: %q", capturedBody.Destination)
			}

			// Verify response fields were mapped correctly.
			if got.ID == nil || *got.ID != 42 {
				t.Fatalf("unexpected ID: %#v", got.ID)
			}
			if got.Type == nil || *got.Type != "pass" {
				t.Fatalf("unexpected type: %#v", got.Type)
			}
			if len(got.Interfaces) != 1 || got.Interfaces[0] != "lan" {
				t.Fatalf("unexpected interfaces: %v", got.Interfaces)
			}
			if got.Description != "allow https" {
				t.Fatalf("unexpected description: %q", got.Description)
			}
			if got.Protocol == nil || *got.Protocol != "tcp" {
				t.Fatalf("unexpected protocol: %#v", got.Protocol)
			}
			if got.Destination != "10.0.0.1" {
				t.Fatalf("unexpected destination: %q", got.Destination)
			}
			if got.DestinationPort == nil || *got.DestinationPort != "443" {
				t.Fatalf("unexpected destination port: %#v", got.DestinationPort)
			}
		})
	}
}

func TestDeleteFirewallRule_Unit(t *testing.T) {
	successResponse := parseDeleteFirewallRuleResponse(t, http.StatusOK, deleteFirewallRuleSuccessJSON)
	unexpectedResponse := parseDeleteFirewallRuleResponse(t, http.StatusBadRequest,
		`{"code":400,"status":"bad request","message":"not found"}`)

	testCases := []struct {
		name      string
		id        int
		response  *DeleteFirewallRuleEndpointResponse
		apiErr    error
		wantError string
	}{
		{
			name:     "returns nil on success",
			id:       42,
			response: successResponse,
		},
		{
			name:      "returns error when api client errors",
			id:        42,
			apiErr:    errors.New("connection refused"),
			wantError: "connection refused",
		},
		{
			name:      "returns error on nil response",
			id:        42,
			response:  nil,
			wantError: "unexpected nil response deleting firewall rule",
		},
		{
			name:      "returns error on unexpected response",
			id:        42,
			response:  unexpectedResponse,
			wantError: "unexpected response deleting firewall rule",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &PFSenseClientV2{
				apiClient: &fakePFSenseAPIClient{
					deleteFirewallRuleEndpointWithResponse: func(ctx context.Context, params *DeleteFirewallRuleEndpointParams, reqEditors ...RequestEditorFn) (*DeleteFirewallRuleEndpointResponse, error) {
						if tc.apiErr != nil {
							return nil, tc.apiErr
						}
						return tc.response, nil
					},
				},
			}

			err := client.DeleteFirewallRule(context.Background(), tc.id)

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
		})
	}
}

func strPtr(s string) *string { return &s }
