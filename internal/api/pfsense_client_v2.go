package pfsense_rest_v2

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
)

type (
	Authorization interface {
		ClientOption() ClientOption
	}
	BasicAuth struct {
		Username string
		Password string
	}
	APIKeyAuth struct {
		APIToken string
	}
)

type PFSenseClientV2 struct {
	url       string
	apiClient *ClientWithResponses
}

type (
	PFSenseBaseConfig struct {
		Hostname string
		Domain   string
	}
)
type PFSenseFirewallRule struct {
	Type            string
	Interfaces      []string
	Disabled        bool
	AddressFamily   string
	Log             bool
	Description     string
	Protocol        string
	Source          string
	SourcePort      string
	Destination     string
	DestinationPort string
}

func NewPFSenseClientV2(url string, auth Authorization, insecure bool) (*PFSenseClientV2, error) {
	apiClient, err := NewClientWithResponses(
		url,
		auth.ClientOption(),
		WithContentTypeJSON,
	)
	if err != nil {
		return nil, err
	} else {
		return &PFSenseClientV2{
			url:       url,
			apiClient: apiClient,
		}, nil
	}
}

func (c *PFSenseClientV2) GetBaseConfig() (*PFSenseBaseConfig, error) {
	response, err := c.apiClient.GetSystemHostnameEndpointWithResponse(context.Background())
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response retrieving base config: %v", response)
	}
	return &PFSenseBaseConfig{
		Hostname: *response.JSON200.Data.Hostname,
		Domain:   *response.JSON200.Data.Domain,
	}, nil
}

func (c *PFSenseClientV2) GetFirewallRules() ([]*PFSenseFirewallRule, error) {
	limit := 0
	response, err := c.apiClient.GetFirewallRulesEndpointWithResponse(
		context.Background(),
		&GetFirewallRulesEndpointParams{
			Limit: &limit,
		},
	)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response retrieving firewall rules: %v", response)
	}
	rulesJSON := response.JSON200.Data

	var rules = []*PFSenseFirewallRule{}
	for _, r := range *rulesJSON {
		rules = append(rules, &PFSenseFirewallRule{
			Type:            string(*r.Type),
			Interfaces:      *r.Interface,
			Disabled:        *r.Disabled,
			AddressFamily:   string(*r.Ipprotocol),
			Log:             *r.Log,
			Description:     *r.Descr,
			Protocol:        string(*r.Protocol),
			Source:          *r.Source,
			SourcePort:      *r.SourcePort,
			Destination:     *r.Destination,
			DestinationPort: *r.DestinationPort,
		})
	}

	return rules, nil
}

func (auth *APIKeyAuth) ClientOption() ClientOption {
	return func(client *Client) error {
		AddHeader(client, "X-API-Key", auth.APIToken)
		return nil
	}
}

func (auth *BasicAuth) ClientOption() ClientOption {
	basicToken := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
	return func(client *Client) error {
		AddHeader(client, "Authorization", "Basic "+basicToken)
		return nil
	}
}

func WithContentTypeJSON(client *Client) error {
	client.RequestEditors = append(client.RequestEditors, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Content-Type", "application/json")
		return nil
	})
	return nil
}

func AddHeader(client *Client, header string, value string) *Client {
	client.RequestEditors = append(client.RequestEditors, func(ctx context.Context, req *http.Request) error {
		req.Header.Add(header, value)
		return nil
	})
	return client
}

func RemoveHeader(client *Client, header string) *Client {
	client.RequestEditors = append(client.RequestEditors, func(ctx context.Context, req *http.Request) error {
		req.Header.Del(header)
		return nil
	})
	return client
}
