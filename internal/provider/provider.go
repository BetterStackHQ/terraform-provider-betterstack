package provider

import (
	"context"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// teamIDValidator checks a resource's own team_id as the provider checks its team_id.
func teamIDValidator() validator.String {
	return stringvalidator.RegexMatches(regexp.MustCompile("^t[1-9][0-9]*$"), idFormMessage("^t[1-9][0-9]*$"))
}

var _ provider.Provider = &BetterStackProvider{}
var _ provider.ProviderWithListResources = &BetterStackProvider{}

type BetterStackProvider struct {
	version   string
	userAgent string
}

// UnifiedAPIVersion is the unified-api version this provider is generated from, which every request
// names in its User-Agent next to the provider's own version: a release's X.Y.Z, or the unified-api
// commit for a build between releases (-ldflags "-X <module>/internal/provider.UnifiedAPIVersion=<commit>").
// Left empty it is the provider's version, which a release shares with its unified-api tag.
var UnifiedAPIVersion = ""

// UserAgent is product/version with the unified-api version as a comment, e.g.
// terraform-provider-betterstack/0.3.0 (unified-api 0.3.0). The apps record both versions on the
// request's span (bs.client.version, bs.unified_api.version).
func UserAgent(product, version string) string {
	unified := UnifiedAPIVersion
	if unified == "" {
		unified = version
	}
	return product + "/" + version + " (unified-api " + unified + ")"
}

func New(version string) func() provider.Provider {
	return NewWithUserAgent(version, UserAgent("terraform-provider-betterstack", version))
}

// NewWithUserAgent lets a bridge (Pulumi) send its own product and version, so API-side traffic
// tells the two apart while the provider itself keeps its default. Build the value with UserAgent.
func NewWithUserAgent(version, userAgent string) func() provider.Provider {
	return func() provider.Provider { return &BetterStackProvider{version: version, userAgent: userAgent} }
}

// The preview APIs' resources, data sources and list resources, which a build with the preview tag
// registers (preview_provider.go). The released provider has none.
var (
	previewResources     []func() resource.Resource
	previewDataSources   []func() datasource.DataSource
	previewListResources []func() list.ListResource
)

type providerModel struct {
	// The settings that fill a templated base URL, such as a Cloud availability zone (provider_variables.go).
	baseURLVariables
	APIToken        types.String `tfsdk:"api_token"`
	Endpoint        types.String `tfsdk:"endpoint"`
	APIRetryMax     types.Int64  `tfsdk:"api_retry_max"`
	APIRetryWaitMin types.Int64  `tfsdk:"api_retry_wait_min"`
	APIRetryWaitMax types.Int64  `tfsdk:"api_retry_wait_max"`
	APITimeout      types.Int64  `tfsdk:"api_timeout"`
	APIRateLimit    types.Int64  `tfsdk:"api_rate_limit"`
	APIRateBurst    types.Int64  `tfsdk:"api_rate_burst"`
	TeamId          types.String `tfsdk:"team_id"`
}

func (p *BetterStackProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "betterstack"
	resp.Version = p.version
}

func (p *BetterStackProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Better Stack resources for uptime monitoring, on-call, telemetry and error tracking.",
		Attributes: map[string]schema.Attribute{
			"api_token":          schema.StringAttribute{Optional: true, Sensitive: true, Description: "Better Stack API token. Falls back to the BETTERSTACK_API_TOKEN environment variable."},
			"endpoint":           schema.StringAttribute{Optional: true, Description: "The API base URL to use instead of Better Stack's, such as a test server."},
			"api_retry_max":      schema.Int64Attribute{Optional: true, Description: "How many times a failed request is retried. Defaults to 4."},
			"api_retry_wait_min": schema.Int64Attribute{Optional: true, Description: "The shortest wait before a retry, in seconds. Defaults to 10."},
			"api_retry_wait_max": schema.Int64Attribute{Optional: true, Description: "The longest wait before a retry, in seconds, also when Better Stack asks to wait longer. Defaults to 300."},
			"api_timeout":        schema.Int64Attribute{Optional: true, Description: "How long one request may take, in seconds. Defaults to 60, and 0 turns the timeout off."},
			"api_rate_limit":     schema.Int64Attribute{Optional: true, Description: "The most requests per second. Defaults to 8, and 0 turns the limit off."},
			"api_rate_burst":     schema.Int64Attribute{Optional: true, Description: "How many requests may go out at once above the rate. Defaults to twice the rate, at least 10, when set to 0."},
			// A credential that can reach several teams says how it picks one; this is that selector,
			// taken from the API description rather than assumed.
			"team_id": schema.StringAttribute{Optional: true, Description: "Selects the team when the API token can reach several.",
				Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile("^t[1-9][0-9]*$"), idFormMessage("^t[1-9][0-9]*$"))}},
		},
	}
	for name, attribute := range baseURLVariableAttributes() {
		resp.Schema.Attributes[name] = attribute
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
	if v := cfg.TeamId.ValueString(); v != "" {
		tenants["team_id"] = v
	}
	client := NewClient(token, cfg.Endpoint.ValueString(), tenants, p.userAgent)
	client.baseURLVars = cfg.baseURLVariables.values()
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
	resp.ListResourceData = client
}

func (p *BetterStackProvider) Resources(ctx context.Context) []func() resource.Resource {
	return append([]func() resource.Resource{
		NewAlertResource,
		NewApplicationResource,
		NewApplicationGroupResource,
		NewAwsLogGroupSubscriptionResource,
		NewCatalogRecordResource,
		NewCatalogRelationResource,
		NewCollectorResource,
		NewCollectorTargetResource,
		NewConnectionResource,
		NewDashboardResource,
		NewDashboardChartResource,
		NewDashboardGroupResource,
		NewDashboardSectionResource,
		NewErrorsJiraIntegrationResource,
		NewErrorsLinearIntegrationResource,
		NewErrorsMicrosoftTeamsIntegrationResource,
		NewErrorsSlackIntegrationResource,
		NewEscalationPolicyResource,
		NewEscalationPolicyGroupResource,
		NewExplorationResource,
		NewExplorationGroupResource,
		NewExtractedMetricResource,
		NewHeartbeatResource,
		NewHeartbeatGroupResource,
		NewIncidentsAmazonSnsIntegrationResource,
		NewIncidentsAwsCloudwatchIntegrationResource,
		NewIncidentsAzureIntegrationResource,
		NewIncidentsDatadogIntegrationResource,
		NewIncidentsElasticIntegrationResource,
		NewIncidentsGoogleMonitoringIntegrationResource,
		NewIncidentsGrafanaIntegrationResource,
		NewIncidentsIncomingEmailIntegrationResource,
		NewIncidentsIncomingWebhookIntegrationResource,
		NewIncidentsJiraIntegrationResource,
		NewIncidentsNewRelicIntegrationResource,
		NewIncidentsOutgoingEmailIntegrationResource,
		NewIncidentsOutgoingPagerdutyIntegrationResource,
		NewIncidentsOutgoingWebhookIntegrationResource,
		NewIncidentsPrometheusIntegrationResource,
		NewIncidentsSplunkOnCallIntegrationResource,
		NewMetadataResource,
		NewMonitorResource,
		NewMonitorGroupResource,
		NewOnCallRotationResource,
		NewOnCallScheduleResource,
		NewReleaseResource,
		NewSeverityResource,
		NewSeverityGroupResource,
		NewSourceResource,
		NewSourceAwsAccountResource,
		NewSourceGcpProjectResource,
		NewSourceGroupResource,
		NewStatusPageResource,
		NewStatusPageGroupResource,
		NewTeamResource,
		NewTeamMemberResource,
	}, previewResources...)
}

// ListResources serve `terraform query`: each finds the objects to import under its parents.
func (p *BetterStackProvider) ListResources(ctx context.Context) []func() list.ListResource {
	return append([]func() list.ListResource{
		NewAlertListResource,
		NewDashboardChartListResource,
		NewDashboardSectionListResource,
	}, previewListResources...)
}

func (p *BetterStackProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return append([]func() datasource.DataSource{
		NewAlertDataSource,
		NewApplicationDataSource,
		NewApplicationGroupDataSource,
		NewAwsLogGroupSubscriptionDataSource,
		NewCatalogRecordDataSource,
		NewCatalogRelationDataSource,
		NewCollectorDataSource,
		NewCollectorTargetDataSource,
		NewConnectionDataSource,
		NewDashboardDataSource,
		NewDashboardChartDataSource,
		NewDashboardGroupDataSource,
		NewDashboardSectionDataSource,
		NewDashboardTemplateDataSource,
		NewErrorsJiraIntegrationDataSource,
		NewErrorsLinearIntegrationDataSource,
		NewErrorsMicrosoftTeamsIntegrationDataSource,
		NewErrorsSlackIntegrationDataSource,
		NewEscalationPolicyDataSource,
		NewEscalationPolicyGroupDataSource,
		NewExplorationDataSource,
		NewExplorationGroupDataSource,
		NewExtractedMetricDataSource,
		NewGcpMetricDataSource,
		NewHeartbeatDataSource,
		NewHeartbeatGroupDataSource,
		NewIncidentsAmazonSnsIntegrationDataSource,
		NewIncidentsAwsCloudwatchIntegrationDataSource,
		NewIncidentsAzureIntegrationDataSource,
		NewIncidentsDatadogIntegrationDataSource,
		NewIncidentsElasticIntegrationDataSource,
		NewIncidentsGoogleMonitoringIntegrationDataSource,
		NewIncidentsGrafanaIntegrationDataSource,
		NewIncidentsIncomingEmailIntegrationDataSource,
		NewIncidentsIncomingWebhookIntegrationDataSource,
		NewIncidentsJiraIntegrationDataSource,
		NewIncidentsNewRelicIntegrationDataSource,
		NewIncidentsOutgoingEmailIntegrationDataSource,
		NewIncidentsOutgoingPagerdutyIntegrationDataSource,
		NewIncidentsOutgoingWebhookIntegrationDataSource,
		NewIncidentsPrometheusIntegrationDataSource,
		NewIncidentsSlackIntegrationDataSource,
		NewIncidentsSplunkOnCallIntegrationDataSource,
		NewIpListDataSource,
		NewMetadataDataSource,
		NewMonitorDataSource,
		NewMonitorGroupDataSource,
		NewOnCallRotationDataSource,
		NewOnCallScheduleDataSource,
		NewReleaseDataSource,
		NewRoleDataSource,
		NewSeverityDataSource,
		NewSeverityGroupDataSource,
		NewSourceDataSource,
		NewSourceGroupDataSource,
		NewStatusPageDataSource,
		NewStatusPageGroupDataSource,
		NewTeamDataSource,
		NewTeamMemberDataSource,
	}, previewDataSources...)
}
