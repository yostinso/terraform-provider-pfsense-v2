// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strings"
	pfsense_rest_v2 "terraform-provider-pfsense-v2/internal/api"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ScaffoldingProvider{}
var _ provider.ProviderWithFunctions = &ScaffoldingProvider{}
var _ provider.ProviderWithEphemeralResources = &ScaffoldingProvider{}

// ScaffoldingProvider defines the provider implementation.
type ScaffoldingProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type ScaffoldingProviderModel struct {
	URL               types.String `tfsdk:"url"`
	Insecure          types.Bool   `tfsdk:"insecure"`
	APIClientUsername types.String `tfsdk:"api_client_username"`
	APIClientPassword types.String `tfsdk:"api_client_password"`
	APIClientToken    types.String `tfsdk:"api_client_token"`
}

func (p *ScaffoldingProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pfsense-v2"
	resp.Version = p.version
}

func (p *ScaffoldingProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The URL of the target pfSense device (e.g. https://192.168.1.1)",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Allow insecure server connections when using SSL",
				Optional:            true,
			},
			"api_client_username": schema.StringAttribute{
				MarkdownDescription: "API client identifier for authentication (i.e. username)",
				Optional:            true,
			},
			"api_client_password": schema.StringAttribute{
				MarkdownDescription: "API client token for authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"api_client_token": schema.StringAttribute{
				MarkdownDescription: "API client token for authentication",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func ConfiguredURL(config *ScaffoldingProviderModel, resp *provider.ConfigureResponse) string {
	// URL from config, can be overridden by env
	const title = "Unknown PFSenseV2 URL"
	const detail = "The provider cannot create the API client as the URL has an Unknown value. " +
		"Either target apply the source of the value first, set the value statically " +
		"in the configuration, or use the PFSENSEV2_URL environment variable."

	if config.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("url"), title, detail)
	}

	url := os.Getenv("PFSENSEV2_URL")
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if url == "" {
		resp.Diagnostics.AddAttributeError(path.Root("url"), title, detail)
	}
	return url
}
func ConfiguredAuth(config *ScaffoldingProviderModel, resp *provider.ConfigureResponse) pfsense_rest_v2.Authorization {
	const title = "No valid PFSenseV2 authentication configured"
	const detail = "One of api_client_username/api_client_password or api_client_token must be set in the provider " +
		"configuration (either with target apply or statically inthe config) or via environment variables " +
		"PFSENSEV2_API_USERNAME, PFSENSEV2_API_PASSWORD, PFSENSE_API_TOKEN."

	if config.APIClientUsername.IsUnknown() && config.APIClientPassword.IsUnknown() && config.APIClientToken.IsUnknown() {
		resp.Diagnostics.AddError(title, detail)
		return nil
	}

	username := os.Getenv("PFSENSEV2_API_USERNAME")
	password := os.Getenv("PFSENSEV2_API_PASSWORD")
	token := os.Getenv("PFSENSEV2_API_TOKEN")
	if !config.APIClientUsername.IsNull() {
		username = config.APIClientUsername.ValueString()
	}
	if !config.APIClientPassword.IsNull() {
		password = config.APIClientPassword.ValueString()
	}
	if !config.APIClientToken.IsNull() {
		token = config.APIClientToken.ValueString()
	}
	if username != "" && password != "" && token == "" {
		resp.Diagnostics.AddError(title, "Only one of api_client_username/api_client_password or api_client_token can be set for authentication.")
		return nil
	}
	if username != "" && password != "" {
		return &pfsense_rest_v2.BasicAuth{
			Username: username,
			Password: password,
		}
	}
	if token != "" {
		return &pfsense_rest_v2.APIKeyAuth{
			APIToken: token,
		}
	}
	resp.Diagnostics.AddError(title, detail)
	return nil
}

func ConfiguredInsecure(config *ScaffoldingProviderModel, resp *provider.ConfigureResponse) bool {
	const title = "Unknown PFSenseV2 Insecure Flag"
	const detail = "The provider cannot create the API client as there is an unknown Insecure flag provided. " +
		"Please check the configuration value or use the PFSENSEV2_INSECURE environment variable."

	if config.Insecure.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("insecure"), title, detail)
	}

	insecure := false

	if len(os.Getenv("PFSENSEV2_INSECURE")) > 0 && strings.ToLower(os.Getenv("PFSENSEV2_INSECURE")) != "false" {
		insecure = true
	}

	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	return insecure
}

func (p *ScaffoldingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ScaffoldingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := ConfiguredURL(&config, resp)
	auth := ConfiguredAuth(&config, resp)
	insecure := ConfiguredInsecure(&config, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	// We now have a valid configuration!
	client, error := pfsense_rest_v2.NewPFSenseClientV2(url, auth, insecure)
	if error != nil {
		resp.Diagnostics.AddError(
			"Unable to Create PFSenseV2 API Client",
			"An unexpected error occurred when creating the PFSenseV2 API client. "+
				error.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ScaffoldingProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *ScaffoldingProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewExampleEphemeralResource,
	}
}

func (p *ScaffoldingProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPFSenseDataSource,
	}
}

func (p *ScaffoldingProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ScaffoldingProvider{
			version: version,
		}
	}
}
