package provider

import (
	"context"
	"fmt"
	"slices"

	pfsense_rest_v2 "terraform-provider-pfsense-v2/internal/api"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &PFSenseDataSource{}

func NewPFSenseDataSource() datasource.DataSource {
	return &PFSenseDataSource{}
}

// PFSenseDataSource defines the data source implementation.
type PFSenseDataSource struct {
	client *pfsense_rest_v2.PFSenseClientV2
}

// PFSenseModel describes the data source data model.
type PFSenseModel struct {
	Hostname      types.String         `tfsdk:"id"`
	FirewallRules PFSenseFirewallRules `tfsdk:"firewall_rules"`
}

type PFSenseFirewallRules []*PFSenseFirewallRule

type PFSenseFirewallRule struct {
	Type            types.String   `tfsdk:"type"`
	Interfaces      []types.String `tfsdk:"interfaces"`
	Disabled        types.Bool     `tfsdk:"disabled"`
	AddressFamily   types.String   `tfsdk:"address_family"`
	Log             types.Bool     `tfsdk:"log"`
	Description     types.String   `tfsdk:"description"`
	Protocol        types.String   `tfsdk:"protocol"`
	Source          types.String   `tfsdk:"source"`
	SourcePort      types.String   `tfsdk:"source_port"`
	Destination     types.String   `tfsdk:"destination"`
	DestinationPort types.String   `tfsdk:"destination_port"`
}

func (rules PFSenseFirewallRules) WANRules() []PFSenseFirewallRule {
	var wanRules []PFSenseFirewallRule
	for _, rule := range rules {
		if slices.Contains(rule.Interfaces, types.StringValue("wan")) {
			wanRules = append(wanRules, *rule)
		}
	}
	return wanRules
}

func (d *PFSenseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configs"
}

func (d *PFSenseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				MarkdownDescription: "PFSense identifier",
				Computed:            true,
			},
			"firewall_rules": schema.ListNestedAttribute{
				MarkdownDescription: "Firewall rules",
				Computed:            true,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"wan_rules": schema.ListNestedAttribute{
							MarkdownDescription: "WAN firewall rules",
							Computed:            true,
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										MarkdownDescription: "Rule type",
										Required:            true,
										Validators: []validator.String{stringvalidator.OneOf(
											string(pfsense_rest_v2.FirewallRuleTypePass),
											string(pfsense_rest_v2.FirewallRuleTypeBlock),
											string(pfsense_rest_v2.FirewallRuleTypeReject),
										)},
									},
									"disabled": schema.BoolAttribute{
										MarkdownDescription: "Whether the rule is disabled",
										Optional:            true,
									},

									"address_family": schema.StringAttribute{
										MarkdownDescription: "Address family (IPv4/IPv6)",
										Optional:            true,
										Validators: []validator.String{stringvalidator.OneOf(
											string(pfsense_rest_v2.FirewallRuleIpprotocolInet),  // IPv4
											string(pfsense_rest_v2.FirewallRuleIpprotocolInet6), // IPv6
										)},
									},
									"log": schema.BoolAttribute{
										MarkdownDescription: "Whether to log packets matching this rule",
										Optional:            true,
									},
									"description": schema.StringAttribute{
										MarkdownDescription: "Rule description",
										Optional:            true,
									},
									"protocol": schema.StringAttribute{
										MarkdownDescription: "Protocol. Supported values: ah, carp, esp, gre, icmp, igmp, ipv6, ospf, pfsync, pim, tcp, tcp/udp, udp.",
										Optional:            true,
										Validators: []validator.String{stringvalidator.OneOf(
											string(pfsense_rest_v2.FirewallRuleProtocolAh),
											string(pfsense_rest_v2.FirewallRuleProtocolCarp),
											string(pfsense_rest_v2.FirewallRuleProtocolEsp),
											string(pfsense_rest_v2.FirewallRuleProtocolGre),
											string(pfsense_rest_v2.FirewallRuleProtocolIcmp),
											string(pfsense_rest_v2.FirewallRuleProtocolIgmp),
											string(pfsense_rest_v2.FirewallRuleProtocolIpv6),
											string(pfsense_rest_v2.FirewallRuleProtocolOspf),
											string(pfsense_rest_v2.FirewallRuleProtocolPfsync),
											string(pfsense_rest_v2.FirewallRuleProtocolPim),
											string(pfsense_rest_v2.FirewallRuleProtocolTcp),
											string(pfsense_rest_v2.FirewallRuleProtocolTcpudp),
											string(pfsense_rest_v2.FirewallRuleProtocolUdp),
										)},
									},
									"source": schema.StringAttribute{
										MarkdownDescription: "The source address this rule applies to. Valid value options are: an existing interface, an IP address, a subnet CIDR, an existing alias, `any`, `(self)`, `l2tp`, `pppoe`. The context of this address can be inverted by prefixing the value with `!`. For interface values, the `:ip` modifier can be appended to the value to use the interface's IP address instead of its entire subnet.",
										Optional:            true,
									},
									"source_port": schema.StringAttribute{
										MarkdownDescription: "The source port this rule applies to. Set to `null` to allow any source port. Valid options are: a TCP/UDP port number, a TCP/UDP port range separated by `:`, an existing port type firewall alias. This field is only available when the following conditions are met: protocol must be one of [ tcp, udp, tcp/udp ].",
										Optional:            true,
										Validators:          []validator.String{PortRangeOrNullValidator{}},
									},
									"destination": schema.StringAttribute{
										MarkdownDescription: "The destination address this rule applies to. Valid value options are: an existing interface, an IP address, a subnet CIDR, an existing alias, `any`, `(self)`, `l2tp`, `pppoe`. The context of this address can be inverted by prefixing the value with `!`. For interface values, the `:ip` modifier can be appended to the value to use the interface's IP address instead of its entire subnet.",
										Optional:            true,
									},
									"destination_port": schema.StringAttribute{
										MarkdownDescription: "The destination port this rule applies to. Set to `null` to allow any destination port. Valid options are: a TCP/UDP port number, a TCP/UDP port range separated by `:`, an existing port type firewall alias. This field is only available when the following conditions are met: protocol must be one of [ tcp, udp, tcp/udp ].",
										Optional:            true,
										Validators:          []validator.String{PortRangeOrNullValidator{}},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *PFSenseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pfsense_rest_v2.PFSenseClientV2)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *pfsense_rest_v2.PFSenseClientV2, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *PFSenseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PFSenseModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }
	baseConfig, err := d.client.GetBaseConfig()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read base config, got error: %s", err))
	}
	firewallRulesResponse, err := d.client.GetFirewallRules()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read firewall rules, got error: %s", err))
	}
	var firewallRules PFSenseFirewallRules
	for _, r := range firewallRulesResponse {
		var ifaces []types.String
		for _, iface := range r.Interfaces {
			ifaces = append(ifaces, types.StringValue(iface))
		}
		firewallRules = append(firewallRules, &PFSenseFirewallRule{
			Type:            types.StringValue(r.Type),
			Interfaces:      ifaces,
			Disabled:        types.BoolValue(r.Disabled),
			AddressFamily:   types.StringValue(r.AddressFamily),
			Log:             types.BoolValue(r.Log),
			Description:     types.StringValue(r.Description),
			Protocol:        types.StringValue(r.Protocol),
			Source:          types.StringValue(r.Source),
			SourcePort:      types.StringValue(r.SourcePort),
			Destination:     types.StringValue(r.Destination),
			DestinationPort: types.StringValue(r.DestinationPort),
		})
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Hostname = types.StringValue(baseConfig.Hostname)
	data.FirewallRules = firewallRules

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
