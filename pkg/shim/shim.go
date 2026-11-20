// Package shim re-exports the plugin-framework provider so external modules (notably the Pulumi
// bridge) can construct it; the implementation itself lives in an internal package.
package shim

import (
	"github.com/betterstackhq/terraform-provider-betterstack/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	pfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
)

// New returns a fresh instance of the Better Stack Terraform plugin-framework provider, sending
// product/version as its User-Agent so the bridge's traffic is distinguishable from the provider's.
func New(product, version string) pfprovider.Provider {
	return provider.NewWithUserAgent(version, product+"/"+version)()
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

// SetPulumiDescriptionEdit rewrites the Terraform-only sentences of the schema's descriptions under
// Pulumi, before the bridge reads them.
func SetPulumiDescriptionEdit(edit func(string) string) { provider.SetPulumiDescriptionEdit(edit) }

// IDInput names a resource's id input for Pulumi, which reserves id: the calendar_id of a rotation.
type IDInput struct{ Resource, Name string }

// IDInputs lists every resource whose id is an input, with the name Pulumi users give it.
func IDInputs() []IDInput {
	return []IDInput{
		{"betterstack_errors_jira_integration", "integration_id"},
		{"betterstack_errors_linear_integration", "integration_id"},
		{"betterstack_errors_microsoft_teams_integration", "integration_id"},
		{"betterstack_errors_slack_integration", "integration_id"},
		{"betterstack_jira_integration", "integration_id"},
		{"betterstack_on_call_rotation", "schedule_id"},
	}
}

// Enum is one input's fixed values, as the API takes them.
type Enum struct {
	Resource, Attribute string
	Values              []string
}

// Enums lists every resource input that takes one of a fixed set of values, for the Pulumi SDKs'
// enum types.
func Enums() []Enum {
	return []Enum{
		{"betterstack_alert", "alert_type", []string{"threshold", "relative", "anomaly_rrcf"}},
		{"betterstack_alert", "anomaly_trigger", []string{"any", "higher", "lower"}},
		{"betterstack_alert", "on_missing_data", []string{"treat_as_zero", "dont_fire", "treat_as_previous", "start_incident"}},
		{"betterstack_alert", "operator", []string{"equal", "not_equal", "higher_than", "higher_than_or_equal", "lower_than", "lower_than_or_equal", "increases_by", "decreases_by", "changes_by"}},
		{"betterstack_alert", "source_mode", []string{"source_variable", "platforms_single_source", "platforms_all_sources"}},
		{"betterstack_amazon_sns_integration", "acknowledged_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_amazon_sns_integration", "resolved_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_amazon_sns_integration", "started_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_application", "platform", []string{"react_errors", "vue_errors", "angular_errors", "react_native_errors", "ember_errors", "solid_errors", "svelte_errors", "sveltekit_errors", "next_js_errors", "nuxt_errors", "remix_errors", "gatsby_errors", "astro_errors", "solidstart_errors", "javascript_errors", "dart_errors", "tanstack_start_react_errors", "react_router_framework_errors", "unity_errors", "android_errors", "ios_errors", "flutter_errors", "asp_dot_net_errors", "asp_dot_net_core_errors", "dot_net_errors", "dot_net_http_errors", "django_errors", "fastapi_errors", "flask_errors", "falcon_errors", "bottle_errors", "asgi_errors", "sanic_errors", "starlette_errors", "tornado_errors", "chalice_errors", "wsgi_errors", "python_errors", "express_errors", "fastify_errors", "nest_js_errors", "koa_errors", "hapi_errors", "node_errors", "bun_errors", "deno_errors", "laravel_errors", "symfony_errors", "php_errors", "spring_errors", "spring_boot_errors", "java_errors", "rails_errors", "rack_middleware_errors", "ruby_errors", "echo_errors", "gin_errors", "fiber_errors", "go_errors", "negroni_errors", "rust_errors", "elixir_errors", "powershell_errors", "rq_errors", "celery_errors", "quart_errors", "pyramid_errors", "connect_errors", "iris_errors", "tryton_errors", "log4j_errors", "logback_errors", "aiohttp_errors", "fasthttp_errors", "dot_net_maui_errors", "unreal_engine_errors", "qt_errors", "electron_errors", "macos_errors", "windows_forms_errors", "wpf_errors", "godot_errors", "minidump_errors", "aws_lambda_dot_net_errors", "aws_lambda_node_errors", "aws_lambda_python_errors", "azure_functions_node_errors", "google_cloud_function_dot_net_errors", "google_cloud_function_node_errors", "google_cloud_function_python_errors", "serverless_python_errors", "cloudflare_pages_errors", "kotlin_errors", "native_errors", "cli_errors"}},
		{"betterstack_catalog_relation", "match_mode", []string{"any", "all"}},
		{"betterstack_collector", "platform", []string{"docker", "swarm", "kubernetes"}},
		{"betterstack_collector_target", "kind", []string{"postgres", "pgbouncer", "mysql", "redis", "mongodb", "memcached", "elasticsearch", "nginx", "apache", "kafka", "prometheus", "traefik"}},
		{"betterstack_collector_target", "scheme", []string{"http", "https"}},
		{"betterstack_collector_target", "ssl_mode", []string{"disable", "require", "verify-ca"}},
		{"betterstack_collector_target", "tls", []string{"false", "true", "skip-verify", "preferred"}},
		{"betterstack_connection", "client_type", []string{"clickhouse", "grafana"}},
		{"betterstack_dashboard_chart", "chart_type", []string{"line_chart", "bar_chart", "pie_chart", "scatter_chart", "gauge_chart", "number_chart", "heatmap_chart", "table_chart", "map_chart", "text_chart", "tail_chart", "static_text_chart", "funnel_chart", "funnel_trend_chart", "anomalies_chart", "hierarchy_chart"}},
		{"betterstack_datadog_integration", "alerting_rule", []string{"alert", "alert_and_warn"}},
		{"betterstack_email_integration", "acknowledged_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_email_integration", "resolved_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_email_integration", "started_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_errors_jira_integration", "occurrence_threshold_interval", []string{"minute", "hour", "day", "week"}},
		{"betterstack_errors_linear_integration", "linear_priority", []string{"no_priority", "urgent", "high", "medium", "low"}},
		{"betterstack_errors_linear_integration", "occurrence_threshold_interval", []string{"minute", "hour", "day", "week"}},
		{"betterstack_errors_microsoft_teams_integration", "occurrence_threshold_interval", []string{"minute", "hour", "day", "week"}},
		{"betterstack_errors_slack_integration", "occurrence_threshold_interval", []string{"minute", "hour", "day", "week"}},
		{"betterstack_errors_slack_integration", "post_updates", []string{"single_channel", "thread_based"}},
		{"betterstack_incoming_webhook_integration", "acknowledged_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incoming_webhook_integration", "resolved_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_incoming_webhook_integration", "started_rule_type", []string{"unused", "any", "all"}},
		{"betterstack_metadata", "mode", []string{"replace", "merge"}},
		{"betterstack_metric", "build_type", []string{"new_data", "historical_logs"}},
		{"betterstack_metric", "type", []string{"string", "string_low_cardinality", "int64_delta", "int64", "uint64_delta", "uint64", "float64_delta", "datetime64_delta", "boolean"}},
		{"betterstack_monitor", "ip_version", []string{"ipv4", "ipv6"}},
		{"betterstack_monitor", "monitor_type", []string{"status", "expected_status_code", "keyword", "keyword_absence", "ping", "tcp", "udp", "smtp", "pop", "imap", "dns", "playwright"}},
		{"betterstack_new_relic_integration", "alerting_rule", []string{"alert", "alert_and_warn"}},
		{"betterstack_on_call_rotation", "rotation_interval", []string{"hour", "day", "week"}},
		{"betterstack_outgoing_webhook_integration", "metadata_api_version", []string{"v2", "v3"}},
		{"betterstack_outgoing_webhook_integration", "trigger_type", []string{"incident_change", "on_call_change", "monitor_change"}},
		{"betterstack_pagerduty_integration", "severity", []string{"info", "warning", "error", "critical"}},
		{"betterstack_source", "platform", []string{"apache2", "aws", "aws_cloudwatch", "aws_ecs", "aws_elb", "aws_fargate", "azure", "azure_logs", "cloudflare_logpush", "cloudflare_prometheus_exporter", "cloudflare_worker", "datadog_agent", "digitalocean", "docker", "dokku", "dotnet", "elasticsearch", "erlang", "fastly", "filebeat", "flights", "fluentbit", "fluentd", "fly_io", "gcp", "go", "google_cloud_pubsub", "haproxy", "heroku", "http", "java", "javascript", "kubernetes", "logstash", "minio", "mongodb", "mysql", "nginx", "open_telemetry", "php", "postgresql", "prometheus", "prometheus_scrape", "python", "rabbitmq", "redis", "render", "rsyslog", "ruby", "syslog-ng", "traefik", "ubuntu", "vector", "vercel_integration", "winlogbeat"}},
		{"betterstack_status_page", "design", []string{"v1", "v2"}},
		{"betterstack_status_page", "layout", []string{"vertical", "horizontal"}},
		{"betterstack_status_page", "theme", []string{"light", "dark", "system"}},
		{"betterstack_status_page_resource", "mark_as_degraded_for", []string{"no_incident", "any_incident", "incident_matching_metadata"}},
		{"betterstack_status_page_resource", "mark_as_down_for", []string{"no_incident", "any_incident", "incident_matching_metadata"}},
		{"betterstack_status_page_resource", "resource_type", []string{"Monitor", "MonitorGroup", "Heartbeat", "HeartbeatGroup", "WebhookIntegration", "EmailIntegration", "IncomingWebhook", "AmazonSnsIntegration", "ManuallyTrackedItem", "LogsChart", "CatalogReference", "ThirdPartyService"}},
		{"betterstack_status_page_resource", "widget_type", []string{"plain", "history", "intraday_history", "response_times", "intraday_response_times", "chart_only"}},
	}
}
