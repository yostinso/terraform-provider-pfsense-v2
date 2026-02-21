package pfsense_rest_v2

import (
	"context"
	"errors"
	"net/http"
	"strings"
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

func TestGetBaseConfig_Unit(t *testing.T) {
	hostname := "pfsense"
	domain := "example.local"

	testCases := []struct {
		name      string
		response  *GetSystemHostnameEndpointResponse
		apiErr    error
		want      *PFSenseBaseConfig
		wantError string
	}{
		{
			name: "returns mapped config on success",
			response: &GetSystemHostnameEndpointResponse{
				JSON200: &struct {
					UnderscoreLinks *map[string]interface{} `json:"_links,omitempty"`
					Code            *int                    `json:"code,omitempty"`
					Data            *SystemHostname         `json:"data,omitempty"`
					Message         *string                 `json:"message,omitempty"`
					ResponseId      *string                 `json:"response_id,omitempty"`
					Status          *string                 `json:"status,omitempty"`
				}{
					Data: &SystemHostname{Hostname: &hostname, Domain: &domain},
				},
				Body: []byte(`{"code":200}`),
				HTTPResponse: &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
				},
			},
			want: &PFSenseBaseConfig{Hostname: "pfsense", Domain: "example.local"},
		},
		{
			name:      "returns error when api client errors",
			apiErr:    errors.New("boom"),
			wantError: "boom",
		},
		{
			name:      "returns error on nil response",
			response:  nil,
			wantError: "unexpected nil response retrieving base config",
		},
		{
			name: "returns error on unexpected response",
			response: &GetSystemHostnameEndpointResponse{
				Body: []byte(`{"code":400}`),
				HTTPResponse: &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
				},
			},
			wantError: "unexpected response retrieving base config",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &PFSenseClientV2{
				apiClient: &fakePFSenseAPIClient{
					getSystemHostnameEndpointWithResponse: func(ctx context.Context, reqEditors ...RequestEditorFn) (*GetSystemHostnameEndpointResponse, error) {
						if tc.apiErr != nil {
							return nil, tc.apiErr
						}
						return tc.response, nil
					},
				},
			}

			got, err := client.GetBaseConfig(context.Background())

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
			if got == nil {
				t.Fatal("expected non-nil config")
			}
			if got.Hostname != tc.want.Hostname {
				t.Fatalf("unexpected hostname: got %q want %q", got.Hostname, tc.want.Hostname)
			}
			if got.Domain != tc.want.Domain {
				t.Fatalf("unexpected domain: got %q want %q", got.Domain, tc.want.Domain)
			}
		})
	}
}
