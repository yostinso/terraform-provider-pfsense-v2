// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"os"

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
				Required:            true,
			},
			"api_client_token": schema.StringAttribute{
				MarkdownDescription: "API client token for authentication",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ScaffoldingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ScaffoldingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate configuration data.
	if config.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown PFSenseV2 URL",
			"The provider cannot create the API client as there is no URL provided. "+
				"Either target apply the source of the value first, set the value statically "+
				"in the configuration, or use the PFSENSEV2_URL environment variable.",
		)
	}
	if config.APIClientUsername.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_client_username"),
			"Unknown PFSenseV2 API Client Username",
			"The provider cannot create the API client as there is no API client username provided. "+
				"Either target apply the source of the value first, set the value statically "+
				"in the configuration, or use the PFSENSEV2_API_USERNAME environment variable.",
		)
	}
	if config.APIClientToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_client_token"),
			"Unknown PFSenseV2 API Client Token",
			"The provider cannot create the API client as there is no API client token provided. "+
				"Either target apply the source of the value first, set the value statically "+
				"in the configuration, or use the PFSENSEV2_API_TOKEN environment variable.",
		)
	}
	if config.Insecure.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("insecure"),
			"Unknown PFSenseV2 Insecure Flag",
			"The provider cannot create the API client as there is an unknown Insecure flag provided. "+
				"Please check the configuration value or use the PFSENSEV2_INSECURE environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Use env variables if provided
	url := os.Getenv("PFSENSEV2_URL")
	username := os.Getenv("PFSENSEV2_API_USERNAME")
	token := os.Getenv("PFSENSEV2_API_TOKEN")
	insecure := false

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if !config.APIClientUsername.IsNull() {
		username = config.APIClientUsername.ValueString()
	}
	if !config.APIClientToken.IsNull() {
		token = config.APIClientToken.ValueString()
	}
	if len(os.Getenv("PFSENSEV2_INSECURE")) > 0 {
		insecure = true
	}

	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing PFSenseV2 URL",
			"The provider cannot create the API client as there is no URL provided. "+
				"Set the URL in the configuration or use the PFSENSEV2_URL environment variable.",
		)
	}
	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_client_username"),
			"Missing PFSenseV2 API Client Username",
			"The provider cannot create the API client as there is no API client username provided. "+
				"Set the API client username in the configuration or use the PFSENSEV2_API_USERNAME environment variable.",
		)
	}
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_client_token"),
			"Missing PFSenseV2 API Client Token",
			"The provider cannot create the API client as there is no API client token provided. "+
				"Set the API client token in the configuration or use the PFSENSEV2_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// We now have a valid configuration!

	client := http.DefaultClient
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
		NewExampleDataSource,
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
