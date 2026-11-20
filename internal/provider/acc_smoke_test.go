package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// The generated smoke matrix: create -> read -> (update ->) import -> destroy for EVERY managed
// resource, against the in-memory mock API built from the same bindings the provider uses. Any
// unresolved unknown after apply, refresh drift, or import mismatch fails the run.
type smokeCase struct {
	name          string
	addr          string
	createConfig  string
	updateConfig  string
	replaceConfig string
	importIgnore  []string
	// importParts is the resource's address as the API spells it: the parent path variables plus
	// the id, which is what a composite `terraform import` string has to carry.
	importParts    []string
	seed           func(*MockAPI) error
	computedFields map[string]map[string]interface{}
	computedChecks map[string]string
}

// importID rebuilds the composite import string out of applied state, so the import step exercises
// the same address the generated ImportState parses.
func importID(addr string, parts []string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return "", fmt.Errorf("%s not found in state", addr)
		}
		values := make([]string, 0, len(parts))
		for _, p := range parts {
			values = append(values, rs.Primary.Attributes[p])
		}
		return strings.Join(values, "/"), nil
	}
}

var smokeCases = []smokeCase{
	{
		name: "amazon_sns_integration",
		addr: "betterstack_amazon_sns_integration.test",
		createConfig: `resource "betterstack_amazon_sns_integration" "test" {
}`,
		updateConfig: `resource "betterstack_amazon_sns_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "application",
		addr: "betterstack_application.test",
		createConfig: `resource "betterstack_application" "test" {
  name                = "example"
  platform            = "react_errors"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_application" "test" {
  name                   = "example"
  platform               = "react_errors"
  github_repository_name = "example-updated"
  deletion_protection    = false
}`,
		replaceConfig: `resource "betterstack_application" "test" {
  name                = "example"
  platform            = "vue_errors"
  deletion_protection = false
}`,
		importIgnore:   []string{"deletion_protection"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "application_group",
		addr: "betterstack_application_group.test",
		createConfig: `resource "betterstack_application_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_application_group" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "aws_cloudwatch_integration",
		addr: "betterstack_aws_cloudwatch_integration.test",
		createConfig: `resource "betterstack_aws_cloudwatch_integration" "test" {
}`,
		updateConfig: `resource "betterstack_aws_cloudwatch_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "azure_integration",
		addr: "betterstack_azure_integration.test",
		createConfig: `resource "betterstack_azure_integration" "test" {
}`,
		updateConfig: `resource "betterstack_azure_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "catalog_attribute",
		addr: "betterstack_catalog_attribute.test",
		createConfig: `resource "betterstack_catalog_attribute" "test" {
  catalog_relation_id = "12345"
  name                = "example"
}`,
		updateConfig: `resource "betterstack_catalog_attribute" "test" {
  catalog_relation_id = "12345"
  name                = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"catalog_relation_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "catalog_record",
		addr: "betterstack_catalog_record.test",
		createConfig: `resource "betterstack_catalog_record" "test" {
  catalog_relation_id = "12345"
  attributes          = [{ attribute = { name = "example" } }]
}`,
		importIgnore:   []string{},
		importParts:    []string{"catalog_relation_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "catalog_relation",
		addr: "betterstack_catalog_relation.test",
		createConfig: `resource "betterstack_catalog_relation" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_catalog_relation" "test" {
  name        = "example"
  description = "example-updated"
}`,
		replaceConfig: `resource "betterstack_catalog_relation" "test" {
  name            = "example"
  organization_id = "12345-updated"
}`,
		importIgnore:   []string{"organization_id"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "chart_alert",
		addr: "betterstack_chart_alert.test",
		createConfig: `resource "betterstack_chart_alert" "test" {
  dashboard_id = 1
  chart_id     = 1
  name         = "example"
  alert_type   = "threshold"
}`,
		updateConfig: `resource "betterstack_chart_alert" "test" {
  dashboard_id = 1
  chart_id     = 1
  name         = "example"
  alert_type   = "threshold"
  string_value = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"dashboard_id", "chart_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "collector",
		addr: "betterstack_collector.test",
		createConfig: `resource "betterstack_collector" "test" {
  name                = "example"
  platform            = "docker"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_collector" "test" {
  name                = "example"
  platform            = "docker"
  note                = "example-updated"
  deletion_protection = false
}`,
		replaceConfig: `resource "betterstack_collector" "test" {
  name                = "example"
  platform            = "swarm"
  deletion_protection = false
}`,
		importIgnore:   []string{"deletion_protection"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "collector_target",
		addr: "betterstack_collector_target.test",
		createConfig: `resource "betterstack_collector_target" "test" {
  collector_id = "12345"
  kind         = "postgres"
  host         = "example"
}`,
		updateConfig: `resource "betterstack_collector_target" "test" {
  collector_id = "12345"
  kind         = "postgres"
  host         = "example"
  service      = "example-updated"
}`,
		replaceConfig: `resource "betterstack_collector_target" "test" {
  collector_id = "12345"
  kind         = "pgbouncer"
  host         = "example"
}`,
		importIgnore:   []string{"api_key", "password"},
		importParts:    []string{"collector_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "connection",
		addr: "betterstack_connection.test",
		createConfig: `resource "betterstack_connection" "test" {
  client_type = "clickhouse"
  team_ids    = ["example"]
}`,
		replaceConfig: `resource "betterstack_connection" "test" {
  client_type = "grafana"
  team_ids    = ["example"]
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "dashboard",
		addr: "betterstack_dashboard.test",
		createConfig: `resource "betterstack_dashboard" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_dashboard" "test" {
  name            = "example"
  date_range_from = "example-updated"
}`,
		importIgnore:   []string{"variables."},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "dashboard_chart",
		addr: "betterstack_dashboard_chart.test",
		createConfig: `resource "betterstack_dashboard_chart" "test" {
  dashboard_id = "12345"
  chart_type   = "line_chart"
  queries      = [{ query_type = "example" }]
}`,
		updateConfig: `resource "betterstack_dashboard_chart" "test" {
  dashboard_id = "12345"
  chart_type   = "line_chart"
  queries      = [{ query_type = "example" }]
  name         = "example-updated"
}`,
		importIgnore: []string{"queries."},
		importParts:  []string{"dashboard_id", "id"},
		computedFields: map[string]map[string]interface{}{
			"queries": {"id": 1},
		},
		computedChecks: map[string]string{
			"queries_result.0.id": "1",
		},
	},
	{
		name: "dashboard_group",
		addr: "betterstack_dashboard_group.test",
		createConfig: `resource "betterstack_dashboard_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_dashboard_group" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "dashboard_section",
		addr: "betterstack_dashboard_section.test",
		createConfig: `resource "betterstack_dashboard_section" "test" {
  dashboard_id = "12345"
  name         = "example"
}`,
		updateConfig: `resource "betterstack_dashboard_section" "test" {
  dashboard_id = "12345"
  name         = "example"
  explanation  = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"dashboard_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "datadog_integration",
		addr: "betterstack_datadog_integration.test",
		createConfig: `resource "betterstack_datadog_integration" "test" {
}`,
		updateConfig: `resource "betterstack_datadog_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "elastic_integration",
		addr: "betterstack_elastic_integration.test",
		createConfig: `resource "betterstack_elastic_integration" "test" {
}`,
		updateConfig: `resource "betterstack_elastic_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "email_integration",
		addr: "betterstack_email_integration.test",
		createConfig: `resource "betterstack_email_integration" "test" {
}`,
		updateConfig: `resource "betterstack_email_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "errors_jira_integration",
		addr: "betterstack_errors_jira_integration.test",
		createConfig: `resource "betterstack_errors_jira_integration" "test" {
  id = "12345"
}`,
		updateConfig: `resource "betterstack_errors_jira_integration" "test" {
  id   = "12345"
  name = "example-updated"
}`,
		importIgnore: []string{},
		importParts:  []string{"id"},
		seed: func(mock *MockAPI) error {
			return mock.Seed("errors_jira_integration", map[string]string{}, "12345")
		},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "errors_linear_integration",
		addr: "betterstack_errors_linear_integration.test",
		createConfig: `resource "betterstack_errors_linear_integration" "test" {
  id = "12345"
}`,
		updateConfig: `resource "betterstack_errors_linear_integration" "test" {
  id   = "12345"
  name = "example-updated"
}`,
		importIgnore: []string{},
		importParts:  []string{"id"},
		seed: func(mock *MockAPI) error {
			return mock.Seed("errors_linear_integration", map[string]string{}, "12345")
		},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "errors_microsoft_teams_integration",
		addr: "betterstack_errors_microsoft_teams_integration.test",
		createConfig: `resource "betterstack_errors_microsoft_teams_integration" "test" {
  id = "12345"
}`,
		updateConfig: `resource "betterstack_errors_microsoft_teams_integration" "test" {
  id                         = "12345"
  occurrence_threshold_count = 2
}`,
		importIgnore: []string{},
		importParts:  []string{"id"},
		seed: func(mock *MockAPI) error {
			return mock.Seed("errors_microsoft_teams_integration", map[string]string{}, "12345")
		},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "errors_slack_integration",
		addr: "betterstack_errors_slack_integration.test",
		createConfig: `resource "betterstack_errors_slack_integration" "test" {
  id = "12345"
}`,
		updateConfig: `resource "betterstack_errors_slack_integration" "test" {
  id                         = "12345"
  occurrence_threshold_count = 2
}`,
		importIgnore: []string{},
		importParts:  []string{"id"},
		seed: func(mock *MockAPI) error {
			return mock.Seed("errors_slack_integration", map[string]string{}, "12345")
		},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "escalation_policy",
		addr: "betterstack_escalation_policy.test",
		createConfig: `resource "betterstack_escalation_policy" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_escalation_policy" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "escalation_policy_group",
		addr: "betterstack_escalation_policy_group.test",
		createConfig: `resource "betterstack_escalation_policy_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_escalation_policy_group" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "exploration",
		addr: "betterstack_exploration.test",
		createConfig: `resource "betterstack_exploration" "test" {
  name    = "example"
  chart   = { chart_type = "line_chart" }
  queries = [{ query_type = "sql_expression" }]
}`,
		updateConfig: `resource "betterstack_exploration" "test" {
  name            = "example"
  chart           = { chart_type = "line_chart" }
  queries         = [{ query_type = "sql_expression" }]
  date_range_from = "example-updated"
}`,
		importIgnore: []string{"queries.", "variables."},
		importParts:  []string{"id"},
		computedFields: map[string]map[string]interface{}{
			"queries": {"id": 1},
		},
		computedChecks: map[string]string{
			"queries_result.0.id": "1",
		},
	},
	{
		name: "exploration_group",
		addr: "betterstack_exploration_group.test",
		createConfig: `resource "betterstack_exploration_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_exploration_group" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "google_monitoring_integration",
		addr: "betterstack_google_monitoring_integration.test",
		createConfig: `resource "betterstack_google_monitoring_integration" "test" {
}`,
		updateConfig: `resource "betterstack_google_monitoring_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "grafana_integration",
		addr: "betterstack_grafana_integration.test",
		createConfig: `resource "betterstack_grafana_integration" "test" {
}`,
		updateConfig: `resource "betterstack_grafana_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "heartbeat",
		addr: "betterstack_heartbeat.test",
		createConfig: `resource "betterstack_heartbeat" "test" {
  name   = "example"
  period = 1
  grace  = 1
}`,
		updateConfig: `resource "betterstack_heartbeat" "test" {
  name      = "example"
  period    = 1
  grace     = 1
  policy_id = "12345-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "heartbeat_group",
		addr: "betterstack_heartbeat_group.test",
		createConfig: `resource "betterstack_heartbeat_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_heartbeat_group" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "incoming_webhook_integration",
		addr: "betterstack_incoming_webhook_integration.test",
		createConfig: `resource "betterstack_incoming_webhook_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incoming_webhook_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "jira_integration",
		addr: "betterstack_jira_integration.test",
		createConfig: `resource "betterstack_jira_integration" "test" {
  id = "12345"
}`,
		updateConfig: `resource "betterstack_jira_integration" "test" {
  id   = "12345"
  name = "example-updated"
}`,
		importIgnore: []string{},
		importParts:  []string{"id"},
		seed: func(mock *MockAPI) error {
			return mock.Seed("jira_integration", map[string]string{}, "12345")
		},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "metadata",
		addr: "betterstack_metadata.test",
		createConfig: `resource "betterstack_metadata" "test" {
  owner_type = "example"
  owner_id   = "12345"
  key        = "example"
  values     = [{ type = "example" }]
}`,
		updateConfig: `resource "betterstack_metadata" "test" {
  owner_type = "example-updated"
  owner_id   = "12345"
  key        = "example"
  values     = [{ type = "example" }]
}`,
		replaceConfig: `resource "betterstack_metadata" "test" {
  owner_type = "example"
  owner_id   = "12345"
  key        = "example"
  values     = [{ type = "example" }]
  mode       = "merge"
}`,
		importIgnore:   []string{"mode"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "metric_expression",
		addr: "betterstack_metric_expression.test",
		createConfig: `resource "betterstack_metric_expression" "test" {
  source_id      = "12345"
  name           = "example"
  sql_expression = "example"
}`,
		updateConfig: `resource "betterstack_metric_expression" "test" {
  source_id      = "12345"
  name           = "example-updated"
  sql_expression = "example"
}`,
		importIgnore:   []string{"build_type"},
		importParts:    []string{"source_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "monitor",
		addr: "betterstack_monitor.test",
		createConfig: `resource "betterstack_monitor" "test" {
  monitor_type = "status"
  url          = "https://example.com"
}`,
		updateConfig: `resource "betterstack_monitor" "test" {
  monitor_type = "status"
  url          = "https://example.com"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "monitor_group",
		addr: "betterstack_monitor_group.test",
		createConfig: `resource "betterstack_monitor_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_monitor_group" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "new_relic_integration",
		addr: "betterstack_new_relic_integration.test",
		createConfig: `resource "betterstack_new_relic_integration" "test" {
}`,
		updateConfig: `resource "betterstack_new_relic_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "on_call_calendar",
		addr: "betterstack_on_call_calendar.test",
		createConfig: `resource "betterstack_on_call_calendar" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_on_call_calendar" "test" {
  name               = "example"
  slack_group_handle = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "on_call_rotation",
		addr: "betterstack_on_call_rotation.test",
		createConfig: `resource "betterstack_on_call_rotation" "test" {
  id                 = "example"
  users              = ["example"]
  rotation_length    = 1
  rotation_interval  = "hour"
  start_rotations_at = "2026-01-01T00:00:00Z"
  end_rotations_at   = "2026-01-01T00:00:00Z"
}`,
		updateConfig: `resource "betterstack_on_call_rotation" "test" {
  id                 = "example"
  users              = ["example"]
  rotation_length    = 1
  rotation_interval  = "hour"
  start_rotations_at = "2026-01-01T00:00:00Z"
  end_rotations_at   = "2026-01-01T00:00:00Z"
  timezone           = "UTC-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "outgoing_email_integration",
		addr: "betterstack_outgoing_email_integration.test",
		createConfig: `resource "betterstack_outgoing_email_integration" "test" {
  recipients = [{ email = "test@example.com" }]
}`,
		updateConfig: `resource "betterstack_outgoing_email_integration" "test" {
  recipients = [{ email = "test@example.com" }]
  name       = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "outgoing_webhook_integration",
		addr: "betterstack_outgoing_webhook_integration.test",
		createConfig: `resource "betterstack_outgoing_webhook_integration" "test" {
  url          = "https://example.com"
  trigger_type = "incident_change"
}`,
		updateConfig: `resource "betterstack_outgoing_webhook_integration" "test" {
  url          = "https://example.com"
  trigger_type = "incident_change"
  name         = "example-updated"
}`,
		replaceConfig: `resource "betterstack_outgoing_webhook_integration" "test" {
  url          = "https://example.com"
  trigger_type = "on_call_change"
}`,
		importIgnore:   []string{"custom_webhook_template_attributes"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "pagerduty_integration",
		addr: "betterstack_pagerduty_integration.test",
		createConfig: `resource "betterstack_pagerduty_integration" "test" {
  name     = "example"
  key      = "example"
  severity = "info"
}`,
		updateConfig: `resource "betterstack_pagerduty_integration" "test" {
  name     = "example-updated"
  key      = "example"
  severity = "info"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "prometheus_integration",
		addr: "betterstack_prometheus_integration.test",
		createConfig: `resource "betterstack_prometheus_integration" "test" {
}`,
		updateConfig: `resource "betterstack_prometheus_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "release",
		addr: "betterstack_release.test",
		createConfig: `resource "betterstack_release" "test" {
  version        = "example"
  application_id = 1
}`,
		replaceConfig: `resource "betterstack_release" "test" {
  version        = "example"
  application_id = 1
  team_id        = "12345-updated"
}`,
		importIgnore:   []string{"environment", "team_id"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "severity",
		addr: "betterstack_severity.test",
		createConfig: `resource "betterstack_severity" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_severity" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "severity_group",
		addr: "betterstack_severity_group.test",
		createConfig: `resource "betterstack_severity_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_severity_group" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "source",
		addr: "betterstack_source.test",
		createConfig: `resource "betterstack_source" "test" {
  name                = "example"
  platform            = "apache2"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_source" "test" {
  name                = "example"
  platform            = "apache2"
  live_tail_pattern   = "example-updated"
  deletion_protection = false
}`,
		replaceConfig: `resource "betterstack_source" "test" {
  name                = "example"
  platform            = "aws"
  deletion_protection = false
}`,
		importIgnore:   []string{"aws_account_id", "aws_external_id", "aws_role_arn", "deletion_protection", "gcp_account_id", "gcp_project_id", "gcp_project_number"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "source_aws_log_group",
		addr: "betterstack_source_aws_log_group.test",
		createConfig: `resource "betterstack_source_aws_log_group" "test" {
  source_id  = "12345"
  region     = "example"
  name       = "example"
  subscribed = true
}`,
		replaceConfig: `resource "betterstack_source_aws_log_group" "test" {
  source_id  = "12345"
  region     = "example-updated"
  name       = "example"
  subscribed = true
}`,
		importIgnore:   []string{},
		importParts:    []string{"source_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "source_group",
		addr: "betterstack_source_group.test",
		createConfig: `resource "betterstack_source_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_source_group" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "splunk_on_call_integration",
		addr: "betterstack_splunk_on_call_integration.test",
		createConfig: `resource "betterstack_splunk_on_call_integration" "test" {
  url = "https://example.com"
}`,
		updateConfig: `resource "betterstack_splunk_on_call_integration" "test" {
  url  = "https://example.com"
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "status_page",
		addr: "betterstack_status_page.test",
		createConfig: `resource "betterstack_status_page" "test" {
  company_name = "example"
  subdomain    = "example-page"
  timezone     = "UTC"
}`,
		updateConfig: `resource "betterstack_status_page" "test" {
  company_name = "example"
  subdomain    = "example-page"
  timezone     = "UTC"
  company_url  = "https://example.com/updated"
}`,
		replaceConfig: `resource "betterstack_status_page" "test" {
  company_name    = "example"
  subdomain       = "example-page"
  timezone        = "UTC"
  organization_id = 2
}`,
		importIgnore:   []string{"password"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "status_page_group",
		addr: "betterstack_status_page_group.test",
		createConfig: `resource "betterstack_status_page_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_status_page_group" "test" {
  name = "example-updated"
}`,
		replaceConfig: `resource "betterstack_status_page_group" "test" {
  name            = "example"
  organization_id = 2
}`,
		importIgnore:   []string{"organization_id"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "status_page_resource",
		addr: "betterstack_status_page_resource.test",
		createConfig: `resource "betterstack_status_page_resource" "test" {
  status_page_id = "12345"
  resource_type  = "Monitor"
  public_name    = "example"
}`,
		updateConfig: `resource "betterstack_status_page_resource" "test" {
  status_page_id = "12345"
  resource_type  = "Monitor"
  public_name    = "example"
  explanation    = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"status_page_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "status_page_section",
		addr: "betterstack_status_page_section.test",
		createConfig: `resource "betterstack_status_page_section" "test" {
  status_page_id = "12345"
}`,
		updateConfig: `resource "betterstack_status_page_section" "test" {
  status_page_id = "12345"
  name           = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"status_page_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name: "team",
		addr: "betterstack_team.test",
		createConfig: `resource "betterstack_team" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_team" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
}

func TestAccSmokeCRUD(t *testing.T) {
	for _, tc := range smokeCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mock := newMockAPI(t)
			if tc.seed != nil {
				if err := tc.seed(mock); err != nil {
					t.Fatal(err)
				}
			}
			mock.InjectItemFields(tc.name, tc.computedFields)
			checks := []resource.TestCheckFunc{resource.TestCheckResourceAttrSet(tc.addr, "id")}
			for path, value := range tc.computedChecks {
				checks = append(checks, resource.TestCheckResourceAttr(tc.addr, path, value))
			}
			steps := []resource.TestStep{
				{
					Config: providerConfig(mock.URL()) + tc.createConfig,
					Check:  resource.ComposeTestCheckFunc(checks...),
				},
			}
			if tc.updateConfig != "" {
				steps = append(steps, resource.TestStep{Config: providerConfig(mock.URL()) + tc.updateConfig})
			}
			if tc.replaceConfig != "" {
				// A create-only field is RequiresReplace; changing it must destroy and recreate
				// the object rather than send a PATCH the API would reject.
				steps = append(steps, resource.TestStep{
					Config: providerConfig(mock.URL()) + tc.replaceConfig,
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(tc.addr, plancheck.ResourceActionReplace)},
					},
				})
			}
			steps = append(steps, resource.TestStep{
				ResourceName:            tc.addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: tc.importIgnore,
				ImportStateIdFunc:       importID(tc.addr, tc.importParts),
			})
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
