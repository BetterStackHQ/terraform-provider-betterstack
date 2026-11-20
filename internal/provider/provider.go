package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &BetterStackProvider{}

type BetterStackProvider struct {
	version   string
	userAgent string
}

func New(version string) func() provider.Provider {
	return NewWithUserAgent(version, "terraform-provider-betterstack/"+version)
}

// NewWithUserAgent lets a bridge (Pulumi) send its own product and version, so API-side traffic
// tells the two apart while the provider itself keeps its default.
func NewWithUserAgent(version, userAgent string) func() provider.Provider {
	return func() provider.Provider { return &BetterStackProvider{version: version, userAgent: userAgent} }
}

type providerModel struct {
	APIToken          types.String `tfsdk:"api_token"`
	Endpoint          types.String `tfsdk:"endpoint"`
	APIRetryMax       types.Int64  `tfsdk:"api_retry_max"`
	APIRetryWaitMin   types.Int64  `tfsdk:"api_retry_wait_min"`
	APIRetryWaitMax   types.Int64  `tfsdk:"api_retry_wait_max"`
	APITimeout        types.Int64  `tfsdk:"api_timeout"`
	APIRateLimit      types.Int64  `tfsdk:"api_rate_limit"`
	APIRateBurst      types.Int64  `tfsdk:"api_rate_burst"`
	BetterStackTeamId types.String `tfsdk:"better_stack_team_id"`
	TeamId            types.String `tfsdk:"team_id"`
	TeamName          types.String `tfsdk:"team_name"`
}

func (p *BetterStackProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "betterstack"
	resp.Version = p.version
}

func (p *BetterStackProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Better Stack (uptime monitoring, on-call, telemetry, error tracking) resources.",
		Attributes: map[string]schema.Attribute{
			"api_token":          schema.StringAttribute{Optional: true, Sensitive: true, Description: "Better Stack API token. Falls back to the BETTERSTACK_API_TOKEN environment variable."},
			"endpoint":           schema.StringAttribute{Optional: true, Description: "Override the API base URL (used in tests)."},
			"api_retry_max":      schema.Int64Attribute{Optional: true, Description: "Maximum retries per request. Default: 4."},
			"api_retry_wait_min": schema.Int64Attribute{Optional: true, Description: "Minimum retry wait in seconds. Default: 10."},
			"api_retry_wait_max": schema.Int64Attribute{Optional: true, Description: "Maximum retry wait in seconds. Default: 300."},
			"api_timeout":        schema.Int64Attribute{Optional: true, Description: "Per-request timeout in seconds; 0 disables it. Default: 60."},
			"api_rate_limit":     schema.Int64Attribute{Optional: true, Description: "Maximum requests per second; 0 disables rate limiting. Default: 8."},
			"api_rate_burst":     schema.Int64Attribute{Optional: true, Description: "Rate limiter burst; 0 chooses twice the rate, at least 10. Default: 0."},
			// A credential that can reach several teams says how it picks one; this is that selector,
			// taken from the API description rather than assumed.
			"better_stack_team_id": schema.StringAttribute{Optional: true, Description: "Selects the team when the API token can reach several."},
			// A credential that can reach several teams says how it picks one; this is that selector,
			// taken from the API description rather than assumed.
			"team_id": schema.StringAttribute{Optional: true, Description: "Selects the team when the API token can reach several."},
			// A credential that can reach several teams says how it picks one; this is that selector,
			// taken from the API description rather than assumed.
			"team_name": schema.StringAttribute{Optional: true, Description: "Selects the team when the API token can reach several."},
		},
	}
}

func (p *BetterStackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	token := cfg.APIToken.ValueString()
	if token == "" {
		token = os.Getenv("BETTERSTACK_API_TOKEN")
	}
	// Public data sources need no token. Authenticated calls reject a missing token before I/O.
	tenants := map[string]string{}
	if v := cfg.BetterStackTeamId.ValueString(); v != "" {
		tenants["better_stack_team_id"] = v
	}
	if v := cfg.TeamId.ValueString(); v != "" {
		tenants["team_id"] = v
	}
	if v := cfg.TeamName.ValueString(); v != "" {
		tenants["team_name"] = v
	}
	client := NewClient(token, cfg.Endpoint.ValueString(), tenants, p.userAgent)
	controls := []struct {
		name     string
		value    types.Int64
		fallback int64
	}{
		{"api_retry_max", cfg.APIRetryMax, 4},
		{"api_retry_wait_min", cfg.APIRetryWaitMin, 10},
		{"api_retry_wait_max", cfg.APIRetryWaitMax, 300},
		{"api_timeout", cfg.APITimeout, 60},
		{"api_rate_limit", cfg.APIRateLimit, 8},
		{"api_rate_burst", cfg.APIRateBurst, 0},
	}
	values := make([]int64, len(controls))
	for i, control := range controls {
		if control.value.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root(control.name), "Unknown API control", "API controls must be known before configuring the provider.")
			return
		}
		values[i] = control.fallback
		if !control.value.IsNull() {
			values[i] = control.value.ValueInt64()
		}
	}
	if err := client.configureControls(values[0], values[1], values[2], values[3], values[4], values[5]); err != nil {
		resp.Diagnostics.AddError("Invalid API controls", err.Error())
		return
	}
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *BetterStackProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAmazonSnsIntegrationResource,
		NewApplicationResource,
		NewApplicationGroupResource,
		NewAwsCloudwatchIntegrationResource,
		NewAzureIntegrationResource,
		NewCatalogAttributeResource,
		NewCatalogRecordResource,
		NewCatalogRelationResource,
		NewChartAlertResource,
		NewCollectorResource,
		NewCollectorTargetResource,
		NewConnectionResource,
		NewDashboardResource,
		NewDashboardChartResource,
		NewDashboardGroupResource,
		NewDashboardSectionResource,
		NewDatadogIntegrationResource,
		NewElasticIntegrationResource,
		NewEmailIntegrationResource,
		NewErrorsJiraIntegrationResource,
		NewErrorsLinearIntegrationResource,
		NewErrorsMicrosoftTeamsIntegrationResource,
		NewErrorsSlackIntegrationResource,
		NewEscalationPolicyResource,
		NewEscalationPolicyGroupResource,
		NewExplorationResource,
		NewExplorationGroupResource,
		NewGoogleMonitoringIntegrationResource,
		NewGrafanaIntegrationResource,
		NewHeartbeatResource,
		NewHeartbeatGroupResource,
		NewIncomingWebhookIntegrationResource,
		NewJiraIntegrationResource,
		NewMetadataResource,
		NewMetricExpressionResource,
		NewMonitorResource,
		NewMonitorGroupResource,
		NewNewRelicIntegrationResource,
		NewOnCallCalendarResource,
		NewOnCallRotationResource,
		NewOutgoingEmailIntegrationResource,
		NewOutgoingWebhookIntegrationResource,
		NewPagerdutyIntegrationResource,
		NewPrometheusIntegrationResource,
		NewReleaseResource,
		NewSeverityResource,
		NewSeverityGroupResource,
		NewSourceResource,
		NewSourceAwsLogGroupResource,
		NewSourceGroupResource,
		NewSplunkOnCallIntegrationResource,
		NewStatusPageResource,
		NewStatusPageGroupResource,
		NewStatusPageResourceResource,
		NewStatusPageSectionResource,
		NewTeamResource,
	}
}

func (p *BetterStackProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAmazonSnsIntegrationDataSource,
		NewApplicationDataSource,
		NewApplicationGroupDataSource,
		NewAwsCloudwatchIntegrationDataSource,
		NewAzureIntegrationDataSource,
		NewCatalogAttributeDataSource,
		NewCatalogRecordDataSource,
		NewCatalogRelationDataSource,
		NewChartAlertDataSource,
		NewCollectorDataSource,
		NewCollectorTargetDataSource,
		NewConnectionDataSource,
		NewDashboardDataSource,
		NewDashboardChartDataSource,
		NewDashboardGroupDataSource,
		NewDashboardSectionDataSource,
		NewDashboardTemplateDataSource,
		NewDatadogIntegrationDataSource,
		NewElasticIntegrationDataSource,
		NewEmailIntegrationDataSource,
		NewErrorsJiraIntegrationDataSource,
		NewErrorsLinearIntegrationDataSource,
		NewErrorsMicrosoftTeamsIntegrationDataSource,
		NewErrorsSlackIntegrationDataSource,
		NewEscalationPolicyDataSource,
		NewEscalationPolicyGroupDataSource,
		NewExplorationDataSource,
		NewExplorationGroupDataSource,
		NewGoogleMonitoringIntegrationDataSource,
		NewGrafanaIntegrationDataSource,
		NewHeartbeatDataSource,
		NewHeartbeatGroupDataSource,
		NewIncomingWebhookIntegrationDataSource,
		NewIpListDataSource,
		NewJiraIntegrationDataSource,
		NewMetadataDataSource,
		NewMetricExpressionDataSource,
		NewMonitorDataSource,
		NewMonitorGroupDataSource,
		NewNewRelicIntegrationDataSource,
		NewOnCallCalendarDataSource,
		NewOnCallRotationDataSource,
		NewOutgoingEmailIntegrationDataSource,
		NewOutgoingWebhookIntegrationDataSource,
		NewPagerdutyIntegrationDataSource,
		NewPrometheusIntegrationDataSource,
		NewReleaseDataSource,
		NewRoleDataSource,
		NewSeverityDataSource,
		NewSeverityGroupDataSource,
		NewSlackIntegrationDataSource,
		NewSourceDataSource,
		NewSourceAwsLogGroupDataSource,
		NewSourceGcpMetricDataSource,
		NewSourceGroupDataSource,
		NewSplunkOnCallIntegrationDataSource,
		NewStatusPageDataSource,
		NewStatusPageGroupDataSource,
		NewStatusPageResourceDataSource,
		NewStatusPageSectionDataSource,
		NewTeamDataSource,
	}
}
