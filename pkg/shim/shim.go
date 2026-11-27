// Package shim re-exports the plugin-framework provider so external modules (notably the Pulumi
// bridge) can construct it; the implementation itself lives in an internal package.
package shim

import (
	"slices"
	"sort"

	"github.com/betterstackhq/terraform-provider-betterstack/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	pfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
)

// What the lists below hold for the preview APIs, which a build with the preview tag fills in
// (preview_shim.go). The released provider has none.
var (
	previewAPIs          []string
	previewIDInputs      []IDInput
	previewClassNames    []ClassName
	previewEnums         []Enum
	previewSingularLists []SingularList
	previewDeletesFirst  []string
)

// APIs lists the Better Stack APIs the provider covers, by their titles in the API description.
func APIs() []string {
	apis := append([]string{"Account", "Errors", "Incidents", "Telemetry"}, previewAPIs...)
	sort.Strings(apis)
	// An API with some resources in preview is in both lists.
	return slices.Compact(apis)
}

// New returns a fresh instance of the Better Stack Terraform plugin-framework provider, sending
// product/version (unified-api <version>) as its User-Agent so the bridge's traffic is
// distinguishable from the provider's.
func New(product, version string) pfprovider.Provider {
	return provider.NewWithUserAgent(version, provider.UserAgent(product, version))()
}

// NewForPulumi is New for the Pulumi bridge: the plan's replacement and deletion-protection
// messages then name properties as names spells them and give Pulumi's fixes.
func NewForPulumi(product, version string, names func(resourceType string, p path.Path) string) pfprovider.Provider {
	provider.UsePulumiMessages(names)
	return New(product, version)
}

// QuoteAPIValues writes API values in Markdown so the Pulumi bridge keeps them as written: a word
// that isn't one of names goes in single quotes, and true, false and null go bare.
func QuoteAPIValues(md string, names map[string]bool) string {
	return provider.QuoteAPIValues(md, names)
}

// QuoteDescriptionValues quotes a schema description's values as QuoteAPIValues does, without escapes.
func QuoteDescriptionValues(md string, names map[string]bool) string {
	return provider.QuoteDescriptionValues(md, names)
}

// SetPulumiDescriptionEdit rewrites the Terraform-only sentences of the schema's descriptions under
// Pulumi, before the bridge reads them.
func SetPulumiDescriptionEdit(edit func(string) string) { provider.SetPulumiDescriptionEdit(edit) }

// SetPulumiDeletesFirst names the resource types Pulumi replaces by deleting first, for the plan's
// replacement warning.
func SetPulumiDeletesFirst(types []string) { provider.SetPulumiDeletesFirst(types) }

// IDInput names a resource's id input for Pulumi, which reserves id: the calendar_id of a rotation.
type IDInput struct{ Resource, Name string }

// IDInputs lists every resource whose id is an input, with the name Pulumi users give it.
func IDInputs() []IDInput {
	return append([]IDInput{
		{"betterstack_errors_jira_integration", "integration_id"},
		{"betterstack_errors_linear_integration", "integration_id"},
		{"betterstack_errors_microsoft_teams_integration", "integration_id"},
		{"betterstack_errors_slack_integration", "integration_id"},
		{"betterstack_incidents_jira_integration", "integration_id"},
		{"betterstack_on_call_rotation", "schedule_id"},
		{"betterstack_source_aws_account", "source_id"},
		{"betterstack_source_gcp_project", "source_id"},
	}, previewIDInputs...)
}

// ClassName is how the SDKs write a resource's name as one word, where splitting the Terraform name
// at its underscores would get a word's capitals wrong: PagerDutyIntegration, not PagerdutyIntegration.
type ClassName struct{ Resource, Name string }

// ClassNames lists every resource and data source whose class name isn't its Terraform name in
// PascalCase.
func ClassNames() []ClassName {
	return append([]ClassName{
		{"betterstack_incidents_aws_cloudwatch_integration", "IncidentsAwsCloudWatchIntegration"},
		{"betterstack_incidents_outgoing_pagerduty_integration", "IncidentsOutgoingPagerDutyIntegration"},
	}, previewClassNames...)
}

// Enum is one input's fixed values, as the API takes them.
type Enum struct {
	Resource, Attribute string
	Values              []string
}

// Enums lists every resource input that takes one of a fixed set of values, for the Pulumi SDKs'
// enum types.
func Enums() []Enum {
	return append([]Enum{
		{"betterstack_alert", "alert_type", []string{"threshold", "relative", "anomaly_rrcf"}},
		{"betterstack_alert", "anomaly_trigger", []string{"any", "higher", "lower"}},
		{"betterstack_alert", "on_missing_data", []string{"treat_as_zero", "dont_fire", "treat_as_previous", "start_incident"}},
		{"betterstack_alert", "operator", []string{"equal", "not_equal", "higher_than", "higher_than_or_equal", "lower_than", "lower_than_or_equal", "increases_by", "decreases_by", "changes_by"}},
		{"betterstack_alert", "source_mode", []string{"source_variable", "platforms_single_source", "platforms_all_sources"}},
		{"betterstack_application", "platform", []string{"aiohttp_errors", "android_errors", "angular_errors", "asgi_errors", "asp_dot_net_core_errors", "asp_dot_net_errors", "astro_errors", "aws_lambda_dot_net_errors", "aws_lambda_node_errors", "aws_lambda_python_errors", "azure_functions_node_errors", "bottle_errors", "bun_errors", "celery_errors", "chalice_errors", "cli_errors", "cloudflare_pages_errors", "connect_errors", "dart_errors", "deno_errors", "django_errors", "dot_net_errors", "dot_net_http_errors", "dot_net_maui_errors", "echo_errors", "electron_errors", "elixir_errors", "ember_errors", "express_errors", "falcon_errors", "fastapi_errors", "fasthttp_errors", "fastify_errors", "fiber_errors", "flask_errors", "flutter_errors", "gatsby_errors", "gin_errors", "go_errors", "godot_errors", "google_cloud_function_dot_net_errors", "google_cloud_function_node_errors", "google_cloud_function_python_errors", "hapi_errors", "ios_errors", "iris_errors", "java_errors", "javascript_errors", "koa_errors", "kotlin_errors", "laravel_errors", "log4j_errors", "logback_errors", "macos_errors", "minidump_errors", "native_errors", "negroni_errors", "nest_js_errors", "next_js_errors", "node_errors", "nuxt_errors", "php_errors", "powershell_errors", "pyramid_errors", "python_errors", "qt_errors", "quart_errors", "rack_middleware_errors", "rails_errors", "react_errors", "react_native_errors", "react_router_framework_errors", "remix_errors", "rq_errors", "ruby_errors", "rust_errors", "sanic_errors", "serverless_python_errors", "solid_errors", "solidstart_errors", "spring_boot_errors", "spring_errors", "starlette_errors", "svelte_errors", "sveltekit_errors", "symfony_errors", "tanstack_start_react_errors", "tornado_errors", "tryton_errors", "unity_errors", "unreal_engine_errors", "vue_errors", "windows_forms_errors", "wpf_errors", "wsgi_errors"}},
		{"betterstack_catalog_relation", "match_mode", []string{"any", "all"}},
		{"betterstack_collector", "platform", []string{"docker", "swarm", "kubernetes"}},
		{"betterstack_collector_target", "kind", []string{"postgres", "pgbouncer", "mysql", "redis", "mongodb", "memcached", "elasticsearch", "nginx", "apache", "kafka", "prometheus", "traefik"}},
		{"betterstack_collector_target", "scheme", []string{"http", "https"}},
		{"betterstack_collector_target", "ssl_mode", []string{"disable", "require", "verify-ca"}},
		{"betterstack_collector_target", "tls", []string{"false", "true", "skip-verify", "preferred"}},
		{"betterstack_connection", "client_type", []string{"clickhouse", "grafana"}},
		{"betterstack_dashboard", "product", []string{"telemetry", "rum"}},
		{"betterstack_dashboard_chart", "chart_type", []string{"line_chart", "bar_chart", "pie_chart", "scatter_chart", "gauge_chart", "number_chart", "heatmap_chart", "table_chart", "map_chart", "text_chart", "tail_chart", "static_text_chart", "funnel_chart", "funnel_trend_chart", "anomalies_chart", "hierarchy_chart"}},
		{"betterstack_errors_jira_integration", "occurrence_threshold_interval", []string{"minute", "hour", "day", "week"}},
		{"betterstack_errors_linear_integration", "linear_priority", []string{"no_priority", "urgent", "high", "medium", "low"}},
		{"betterstack_errors_linear_integration", "occurrence_threshold_interval", []string{"minute", "hour", "day", "week"}},
		{"betterstack_errors_microsoft_teams_integration", "occurrence_threshold_interval", []string{"minute", "hour", "day", "week"}},
		{"betterstack_errors_slack_integration", "occurrence_threshold_interval", []string{"minute", "hour", "day", "week"}},
		{"betterstack_errors_slack_integration", "post_updates", []string{"single_channel", "thread_based"}},
		{"betterstack_exploration", "product", []string{"telemetry", "rum"}},
		{"betterstack_exploration", "source_type", []string{"web_events", "logs", "spans"}},
		{"betterstack_extracted_metric", "build_type", []string{"new_data", "historical_logs"}},
		{"betterstack_extracted_metric", "type", []string{"string", "string_low_cardinality", "int64_delta", "int64", "uint64_delta", "uint64", "float64_delta", "datetime64_delta", "boolean"}},
		{"betterstack_incidents_amazon_sns_integration", "acknowledged_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_amazon_sns_integration", "resolved_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_amazon_sns_integration", "started_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_datadog_integration", "alerting_rule", []string{"alert", "alert_and_warn"}},
		{"betterstack_incidents_incoming_email_integration", "acknowledged_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_incoming_email_integration", "resolved_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_incoming_email_integration", "started_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_incoming_webhook_integration", "acknowledged_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_incoming_webhook_integration", "resolved_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_incoming_webhook_integration", "started_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incidents_new_relic_integration", "alerting_rule", []string{"alert", "alert_and_warn"}},
		{"betterstack_incidents_outgoing_pagerduty_integration", "severity", []string{"info", "warning", "error", "critical"}},
		{"betterstack_incidents_outgoing_webhook_integration", "metadata_api_version", []string{"v2", "v3"}},
		{"betterstack_incidents_outgoing_webhook_integration", "trigger_type", []string{"incident_change", "on_call_change", "monitor_change"}},
		{"betterstack_metadata", "mode", []string{"replace", "merge"}},
		{"betterstack_monitor", "ip_version", []string{"ipv4", "ipv6"}},
		{"betterstack_monitor", "monitor_type", []string{"status", "expected_status_code", "keyword", "keyword_absence", "ping", "tcp", "udp", "smtp", "pop", "imap", "dns", "playwright"}},
		{"betterstack_on_call_rotation", "rotation_interval", []string{"hour", "day", "week"}},
		{"betterstack_source", "platform", []string{"apache2", "aws", "aws_cloudwatch", "aws_ecs", "aws_elb", "aws_fargate", "azure", "azure_logs", "claude_platform", "cloudflare_logpush", "cloudflare_prometheus_exporter", "cloudflare_worker", "datadog_agent", "digitalocean", "docker", "dokku", "dotnet", "elasticsearch", "erlang", "fastly", "filebeat", "flights", "fluentbit", "fluentd", "fly_io", "gcp", "go", "google_cloud_pubsub", "haproxy", "heroku", "http", "java", "javascript", "kubernetes", "logstash", "minio", "mongodb", "mysql", "nginx", "open_telemetry", "oracle_cloud", "php", "postgresql", "prometheus", "prometheus_scrape", "python", "rabbitmq", "redis", "render", "rsyslog", "ruby", "syslog-ng", "traefik", "ubuntu", "vector", "vercel_integration", "winlogbeat"}},
		{"betterstack_status_page", "design", []string{"v1", "v2"}},
		{"betterstack_status_page", "layout", []string{"vertical", "horizontal"}},
		{"betterstack_status_page", "theme", []string{"light", "dark", "system"}},
		{"betterstack_team_member", "role", []string{"admin", "billing_admin", "team_lead", "responder", "member", "custom"}},
	}, previewEnums...)
}

// SingularList is a list attribute, as field or field.subfield, whose Terraform name already names
// the whole list, so Pulumi keeps it as it is: ipAllowlist, not ipAllowlists.
type SingularList struct{ Resource, Path string }

// SingularLists lists every such attribute of the resources and their data sources.
func SingularLists() []SingularList {
	return append([]SingularList{
		{"betterstack_alert", "additional_conditions.series_names_except"},
		{"betterstack_alert", "series_names_except"},
		{"betterstack_connection", "ip_allowlist"},
		{"betterstack_dashboard_chart", "queries.group_by"},
		{"betterstack_dashboard_chart", "queries.having"},
		{"betterstack_dashboard_chart", "queries.order_by"},
		{"betterstack_dashboard_chart", "queries.point_size"},
		{"betterstack_dashboard_chart", "queries.where"},
		{"betterstack_dashboard_chart", "settings.series_order"},
		{"betterstack_exploration", "chart.settings.series_order"},
		{"betterstack_exploration", "queries.group_by"},
		{"betterstack_exploration", "queries.having"},
		{"betterstack_exploration", "queries.order_by"},
		{"betterstack_exploration", "queries.point_size"},
		{"betterstack_exploration", "queries.where"},
		{"betterstack_incidents_outgoing_webhook_integration", "custom_webhook_template.headers_template"},
		{"betterstack_incidents_outgoing_webhook_integration", "custom_webhook_template_attributes.headers_template"},
		{"betterstack_status_page", "ip_allowlist"},
		{"betterstack_team", "incident_grouping_metadata_scope"},
	}, previewSingularLists...)
}

// DeletesFirst lists the resources whose replacement Pulumi must start with the delete: the new object
// would take a value only one object may hold (a status page's subdomain), which the
// old one still holds, so creating first would fail every replacement.
func DeletesFirst() []string {
	return append([]string{"betterstack_status_page"}, previewDeletesFirst...)
}
