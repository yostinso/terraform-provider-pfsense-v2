package pfsense_rest_v2

import "context"

// fakePFSenseAPIClient is a test double for pfsenseAPIClient.
// Each field is a function that can be set per-test to control behaviour.
// Nil fields return (nil, nil) by default, mimicking an unused stub method.
type fakePFSenseAPIClient struct {
	getSystemHostnameEndpointWithResponse  func(ctx context.Context, reqEditors ...RequestEditorFn) (*GetSystemHostnameEndpointResponse, error)
	getFirewallRulesEndpointWithResponse   func(ctx context.Context, params *GetFirewallRulesEndpointParams, reqEditors ...RequestEditorFn) (*GetFirewallRulesEndpointResponse, error)
	getFirewallRuleEndpointWithResponse    func(ctx context.Context, params *GetFirewallRuleEndpointParams, reqEditors ...RequestEditorFn) (*GetFirewallRuleEndpointResponse, error)
	postFirewallRuleEndpointWithResponse   func(ctx context.Context, body PostFirewallRuleEndpointJSONRequestBody, reqEditors ...RequestEditorFn) (*PostFirewallRuleEndpointResponse, error)
	deleteFirewallRuleEndpointWithResponse func(ctx context.Context, params *DeleteFirewallRuleEndpointParams, reqEditors ...RequestEditorFn) (*DeleteFirewallRuleEndpointResponse, error)
}

func (f *fakePFSenseAPIClient) GetSystemHostnameEndpointWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetSystemHostnameEndpointResponse, error) {
	if f.getSystemHostnameEndpointWithResponse == nil {
		return nil, nil
	}
	return f.getSystemHostnameEndpointWithResponse(ctx, reqEditors...)
}

func (f *fakePFSenseAPIClient) GetFirewallRulesEndpointWithResponse(ctx context.Context, params *GetFirewallRulesEndpointParams, reqEditors ...RequestEditorFn) (*GetFirewallRulesEndpointResponse, error) {
	if f.getFirewallRulesEndpointWithResponse == nil {
		return nil, nil
	}
	return f.getFirewallRulesEndpointWithResponse(ctx, params, reqEditors...)
}

func (f *fakePFSenseAPIClient) GetFirewallRuleEndpointWithResponse(ctx context.Context, params *GetFirewallRuleEndpointParams, reqEditors ...RequestEditorFn) (*GetFirewallRuleEndpointResponse, error) {
	if f.getFirewallRuleEndpointWithResponse == nil {
		return nil, nil
	}
	return f.getFirewallRuleEndpointWithResponse(ctx, params, reqEditors...)
}

func (f *fakePFSenseAPIClient) PostFirewallRuleEndpointWithResponse(ctx context.Context, body PostFirewallRuleEndpointJSONRequestBody, reqEditors ...RequestEditorFn) (*PostFirewallRuleEndpointResponse, error) {
	if f.postFirewallRuleEndpointWithResponse == nil {
		return nil, nil
	}
	return f.postFirewallRuleEndpointWithResponse(ctx, body, reqEditors...)
}

func (f *fakePFSenseAPIClient) DeleteFirewallRuleEndpointWithResponse(ctx context.Context, params *DeleteFirewallRuleEndpointParams, reqEditors ...RequestEditorFn) (*DeleteFirewallRuleEndpointResponse, error) {
	if f.deleteFirewallRuleEndpointWithResponse == nil {
		return nil, nil
	}
	return f.deleteFirewallRuleEndpointWithResponse(ctx, params, reqEditors...)
}
