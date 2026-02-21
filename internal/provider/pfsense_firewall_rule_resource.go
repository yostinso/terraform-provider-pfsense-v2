package provider

import (
	"context"
	"fmt"
	"strconv"

	pfsense_rest_v2 "terraform-provider-pfsense-v2/internal/api"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &PFSenseFirewallRuleResource{}

func NewPFSenseFirewallRuleResource() resource.Resource {
	return &PFSenseFirewallRuleResource{}
}

type PFSenseFirewallRuleResource struct {
	client *pfsense_rest_v2.PFSenseClientV2
}

// PFSenseFirewallRuleResourceModel is the Terraform state model for a managed firewall rule.
type PFSenseFirewallRuleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Type            types.String `tfsdk:"type"`
	Interfaces      types.List   `tfsdk:"interfaces"`
	Disabled        types.Bool   `tfsdk:"disabled"`
	AddressFamily   types.String `tfsdk:"address_family"`
	Log             types.Bool   `tfsdk:"log"`
	Description     types.String `tfsdk:"description"`
	Protocol        types.String `tfsdk:"protocol"`
	Source          types.String `tfsdk:"source"`
	SourcePort      types.String `tfsdk:"source_port"`
	Destination     types.String `tfsdk:"destination"`
	DestinationPort types.String `tfsdk:"destination_port"`
}

func (r *PFSenseFirewallRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rule"
}

func (r *PFSenseFirewallRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a pfSense firewall rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal pfSense rule ID assigned on creation.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Rule action: pass, block, or reject.",
				Validators: []validator.String{stringvalidator.OneOf(
					string(pfsense_rest_v2.FirewallRuleTypePass),
					string(pfsense_rest_v2.FirewallRuleTypeBlock),
					string(pfsense_rest_v2.FirewallRuleTypeReject),
				)},
			},
			"interfaces": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Interfaces this rule applies to.",
			},
			"disabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the rule is disabled.",
			},
			"address_family": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IP version: inet (IPv4), inet6 (IPv6), or inet46 (both).",
				Validators: []validator.String{stringvalidator.OneOf(
					string(pfsense_rest_v2.FirewallRuleIpprotocolInet),
					string(pfsense_rest_v2.FirewallRuleIpprotocolInet6),
					string(pfsense_rest_v2.FirewallRuleIpprotocolInet46),
				)},
			},
			"log": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to log packets matching this rule.",
			},
			"description": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable description for this rule.",
			},
			"protocol": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IP/transport protocol. Supported values: ah, carp, esp, gre, icmp, igmp, ipv6, ospf, pfsync, pim, tcp, tcp/udp, udp.",
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
				Required:            true,
				MarkdownDescription: "Source address (interface name, IP, CIDR, alias, 'any', etc.).",
			},
			"source_port": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Source port or range (only valid for tcp, udp, tcp/udp).",
				Validators:          []validator.String{PortRangeOrNullValidator{}},
			},
			"destination": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Destination address (interface name, IP, CIDR, alias, 'any', etc.).",
			},
			"destination_port": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Destination port or range (only valid for tcp, udp, tcp/udp).",
				Validators:          []validator.String{PortRangeOrNullValidator{}},
			},
		},
	}
}

func (r *PFSenseFirewallRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*pfsense_rest_v2.PFSenseClientV2)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *pfsense_rest_v2.PFSenseClientV2, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *PFSenseFirewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PFSenseFirewallRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiRule := planToAPIRule(ctx, &plan, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateFirewallRule(ctx, apiRule)
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall rule", err.Error())
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Created firewall rule with ID %d", *created.ID))

	apiRuleToState(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PFSenseFirewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PFSenseFirewallRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID", fmt.Sprintf("Cannot parse firewall rule ID %q: %s", state.ID.ValueString(), err))
		return
	}

	rule, err := r.client.GetFirewallRule(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall rule", err.Error())
		return
	}

	apiRuleToState(rule, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not yet implemented; in-place updates will trigger a replace via plan modifiers.
func (r *PFSenseFirewallRuleResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Firewall rule updates are not yet implemented; remove and re-add the rule to change it.")
}

func (r *PFSenseFirewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PFSenseFirewallRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid resource ID", fmt.Sprintf("Cannot parse firewall rule ID %q: %s", state.ID.ValueString(), err))
		return
	}

	if err := r.client.DeleteFirewallRule(ctx, id); err != nil {
		resp.Diagnostics.AddError("Error deleting firewall rule", err.Error())
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Deleted firewall rule with ID %d", id))
}

// planToAPIRule converts the Terraform plan/state model into the API input struct.
func planToAPIRule(ctx context.Context, m *PFSenseFirewallRuleResourceModel, diags interface{ HasError() bool }) *pfsense_rest_v2.PFSenseFirewallRule {
	var ifaces []string
	m.Interfaces.ElementsAs(ctx, &ifaces, false)

	ruleType := m.Type.ValueString()
	proto := m.Protocol.ValueStringPointer()
	srcPort := m.SourcePort.ValueStringPointer()
	dstPort := m.DestinationPort.ValueStringPointer()

	// Treat empty optional strings as nil so the API doesn't receive empty values.
	if proto != nil && *proto == "" {
		proto = nil
	}
	if srcPort != nil && *srcPort == "" {
		srcPort = nil
	}
	if dstPort != nil && *dstPort == "" {
		dstPort = nil
	}

	return &pfsense_rest_v2.PFSenseFirewallRule{
		Type:            &ruleType,
		Interfaces:      ifaces,
		Disabled:        m.Disabled.ValueBool(),
		AddressFamily:   m.AddressFamily.ValueString(),
		Log:             m.Log.ValueBool(),
		Description:     m.Description.ValueString(),
		Protocol:        proto,
		Source:          m.Source.ValueString(),
		SourcePort:      srcPort,
		Destination:     m.Destination.ValueString(),
		DestinationPort: dstPort,
	}
}

// apiRuleToState writes the API response fields back into the Terraform state model.
func apiRuleToState(rule *pfsense_rest_v2.PFSenseFirewallRule, m *PFSenseFirewallRuleResourceModel) {
	m.ID = types.StringValue(strconv.Itoa(*rule.ID))
	m.Type = types.StringPointerValue(rule.Type)
	m.Disabled = types.BoolValue(rule.Disabled)
	m.AddressFamily = types.StringValue(rule.AddressFamily)
	m.Log = types.BoolValue(rule.Log)
	m.Description = types.StringValue(rule.Description)
	m.Protocol = types.StringPointerValue(rule.Protocol)
	m.Source = types.StringValue(rule.Source)
	m.SourcePort = types.StringPointerValue(rule.SourcePort)
	m.Destination = types.StringValue(rule.Destination)
	m.DestinationPort = types.StringPointerValue(rule.DestinationPort)

	ifaces := make([]types.String, len(rule.Interfaces))
	for i, iface := range rule.Interfaces {
		ifaces[i] = types.StringValue(iface)
	}
	m.Interfaces, _ = types.ListValueFrom(context.Background(), types.StringType, ifaces)
}
