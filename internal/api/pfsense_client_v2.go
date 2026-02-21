package pfsense_rest_v2

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	apiClient pfsenseAPIClient
}

type pfsenseAPIClient interface {
	GetSystemHostnameEndpointWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetSystemHostnameEndpointResponse, error)
	GetFirewallRulesEndpointWithResponse(ctx context.Context, params *GetFirewallRulesEndpointParams, reqEditors ...RequestEditorFn) (*GetFirewallRulesEndpointResponse, error)
	GetFirewallRuleEndpointWithResponse(ctx context.Context, params *GetFirewallRuleEndpointParams, reqEditors ...RequestEditorFn) (*GetFirewallRuleEndpointResponse, error)
	PostFirewallRuleEndpointWithResponse(ctx context.Context, body PostFirewallRuleEndpointJSONRequestBody, reqEditors ...RequestEditorFn) (*PostFirewallRuleEndpointResponse, error)
	DeleteFirewallRuleEndpointWithResponse(ctx context.Context, params *DeleteFirewallRuleEndpointParams, reqEditors ...RequestEditorFn) (*DeleteFirewallRuleEndpointResponse, error)
}

type (
	PFSenseBaseConfig struct {
		Hostname string
		Domain   string
	}
)
type PFSenseFirewallRule struct {
	ID              *int
	Type            *string
	Interfaces      []string
	Disabled        bool
	AddressFamily   string
	Log             bool
	Description     string
	Protocol        *string
	Source          string
	SourcePort      *string
	Destination     string
	DestinationPort *string
}

func NewPFSenseClientV2(url string, auth Authorization, insecure bool) (*PFSenseClientV2, error) {
	httpClient := &http.Client{}
	parsedUrl, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}
	if insecure && parsedUrl.Scheme == "https" {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig.InsecureSkipVerify = true
		httpClient.Transport = transport
	}

	apiClient, err := NewClientWithResponses(
		url,
		auth.ClientOption(),
		WithHTTPClient(httpClient),
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

func (c *PFSenseClientV2) URL() string {
	return c.url
}

func (c *PFSenseClientV2) GetBaseConfig(ctx context.Context) (*PFSenseBaseConfig, error) {
	response, err := c.apiClient.GetSystemHostnameEndpointWithResponse(ctx)
	if err != nil {
		tflog.Debug(ctx, "GetSystemHostnameEndpointWithResponse"+FormatResponse(ctx, nil, nil, err))
		return nil, err
	}
	if response == nil {
		tflog.Debug(ctx, "GetSystemHostnameEndpointWithResponse nil response")
		return nil, fmt.Errorf("unexpected nil response retrieving base config")
	}
	tflog.Debug(ctx, "GetSystemHostnameEndpointWithResponse"+FormatResponse(ctx, response.Body, response.HTTPResponse, err))
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response retrieving base config: %v", response)
	}
	return &PFSenseBaseConfig{
		Hostname: *response.JSON200.Data.Hostname,
		Domain:   *response.JSON200.Data.Domain,
	}, nil
}

func (c *PFSenseClientV2) GetFirewallRules(ctx context.Context) ([]*PFSenseFirewallRule, error) {
	limit := 0
	response, err := c.apiClient.GetFirewallRulesEndpointWithResponse(
		ctx,
		&GetFirewallRulesEndpointParams{
			Limit: &limit,
		},
	)
	if err != nil {
		tflog.Debug(ctx, "GetFirewallRulesEndpointWithResponse"+FormatResponse(ctx, nil, nil, err))
		return nil, err
	}

	if response == nil {
		tflog.Debug(ctx, "GetFirewallRulesEndpointWithResponse nil response")
		return nil, fmt.Errorf("unexpected nil response retrieving firewall rules")
	}
	tflog.Debug(ctx, "GetFirewallRulesEndpointWithResponse"+FormatResponse(ctx, response.Body, response.HTTPResponse, err))

	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected response retrieving firewall rules: %v", response)
	}
	rulesJSON := response.JSON200.Data

	var rules = []*PFSenseFirewallRule{}
	for _, r := range *rulesJSON {

		rules = append(rules, &PFSenseFirewallRule{
			ID:              r.Id,
			Type:            (*string)(r.Type),
			Interfaces:      *r.Interface,
			Disabled:        *r.Disabled,
			AddressFamily:   string(*r.Ipprotocol),
			Log:             *r.Log,
			Description:     string(*r.Descr),
			Protocol:        (*string)(r.Protocol),
			Source:          *r.Source,
			SourcePort:      r.SourcePort,
			Destination:     *r.Destination,
			DestinationPort: r.DestinationPort,
		})
	}

	return rules, nil
}

func (c *PFSenseClientV2) CreateFirewallRule(ctx context.Context, rule *PFSenseFirewallRule) (*PFSenseFirewallRule, error) {
	body := PostFirewallRuleEndpointJSONRequestBody{
		Type:            PostFirewallRuleEndpointJSONBodyType(*rule.Type),
		Interface:       rule.Interfaces,
		Ipprotocol:      PostFirewallRuleEndpointJSONBodyIpprotocol(rule.AddressFamily),
		Source:          rule.Source,
		Destination:     rule.Destination,
		Descr:           &rule.Description,
		Protocol:        (*PostFirewallRuleEndpointJSONBodyProtocol)(rule.Protocol),
		Disabled:        &rule.Disabled,
		Log:             &rule.Log,
		SourcePort:      rule.SourcePort,
		DestinationPort: rule.DestinationPort,
	}
	response, err := c.apiClient.PostFirewallRuleEndpointWithResponse(ctx, body)
	if err != nil {
		tflog.Debug(ctx, "PostFirewallRuleEndpointWithResponse"+FormatResponse(ctx, nil, nil, err))
		return nil, err
	}
	if response == nil {
		tflog.Debug(ctx, "PostFirewallRuleEndpointWithResponse nil response")
		return nil, fmt.Errorf("unexpected nil response creating firewall rule")
	}
	tflog.Debug(ctx, "PostFirewallRuleEndpointWithResponse"+FormatResponse(ctx, response.Body, response.HTTPResponse, err))
	if response.JSON200 == nil || response.JSON200.Data == nil {
		return nil, fmt.Errorf("unexpected response creating firewall rule: %v", response)
	}
	data := response.JSON200.Data
	return &PFSenseFirewallRule{
		ID:              data.Id,
		Type:            (*string)(data.Type),
		Interfaces:      *data.Interface,
		Disabled:        *data.Disabled,
		AddressFamily:   string(*data.Ipprotocol),
		Log:             *data.Log,
		Description:     *data.Descr,
		Protocol:        (*string)(data.Protocol),
		Source:          *data.Source,
		SourcePort:      data.SourcePort,
		Destination:     *data.Destination,
		DestinationPort: data.DestinationPort,
	}, nil
}

func (c *PFSenseClientV2) GetFirewallRule(ctx context.Context, id int) (*PFSenseFirewallRule, error) {
	setIDParam := func(ctx context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("id", strconv.Itoa(id))
		req.URL.RawQuery = q.Encode()
		return nil
	}
	response, err := c.apiClient.GetFirewallRuleEndpointWithResponse(ctx, nil, setIDParam)
	if err != nil {
		tflog.Debug(ctx, "GetFirewallRuleEndpointWithResponse"+FormatResponse(ctx, nil, nil, err))
		return nil, err
	}
	if response == nil {
		tflog.Debug(ctx, "GetFirewallRuleEndpointWithResponse nil response")
		return nil, fmt.Errorf("unexpected nil response getting firewall rule")
	}
	tflog.Debug(ctx, "GetFirewallRuleEndpointWithResponse"+FormatResponse(ctx, response.Body, response.HTTPResponse, err))
	if response.JSON200 == nil || response.JSON200.Data == nil {
		return nil, fmt.Errorf("unexpected response getting firewall rule: %v", response)
	}
	data := response.JSON200.Data
	return &PFSenseFirewallRule{
		ID:              data.Id,
		Type:            (*string)(data.Type),
		Interfaces:      *data.Interface,
		Disabled:        *data.Disabled,
		AddressFamily:   string(*data.Ipprotocol),
		Log:             *data.Log,
		Description:     *data.Descr,
		Protocol:        (*string)(data.Protocol),
		Source:          *data.Source,
		SourcePort:      data.SourcePort,
		Destination:     *data.Destination,
		DestinationPort: data.DestinationPort,
	}, nil
}

func (c *PFSenseClientV2) DeleteFirewallRule(ctx context.Context, id int) error {
	setIDParam := func(ctx context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("id", strconv.Itoa(id))
		req.URL.RawQuery = q.Encode()
		return nil
	}
	response, err := c.apiClient.DeleteFirewallRuleEndpointWithResponse(ctx, nil, setIDParam)
	if err != nil {
		tflog.Debug(ctx, "DeleteFirewallRuleEndpointWithResponse"+FormatResponse(ctx, nil, nil, err))
		return err
	}
	if response == nil {
		tflog.Debug(ctx, "DeleteFirewallRuleEndpointWithResponse nil response")
		return fmt.Errorf("unexpected nil response deleting firewall rule")
	}
	tflog.Debug(ctx, "DeleteFirewallRuleEndpointWithResponse"+FormatResponse(ctx, response.Body, response.HTTPResponse, err))
	if response.JSON200 == nil {
		return fmt.Errorf("unexpected response deleting firewall rule: %v", response)
	}
	return nil
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

type ResponseWithBodyAndStatus struct {
	Body         []byte
	HTTPResponse *http.Response
}

func FormatResponse(ctx context.Context, body []byte, httpResponse *http.Response, err error) string {
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	if body == nil {
		return "nil response"
	}
	return fmt.Sprintf("Status: %s, Body: %s", httpResponse.Status, string(body))
}
