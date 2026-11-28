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
	name string
	addr string
	// idKey is the attribute holding the object's identity: `id`, or the resource's id_field.
	idKey         string
	createConfig  string
	updateConfig  string
	replaceConfig string
	// inPlaceConfigs each change fields the API changes through an action of their own.
	inPlaceConfigs []string
	importIgnore   []string
	// importParts is the resource's address as the API spells it: the parent path variables plus
	// the id, which is what a composite `terraform import` string has to carry.
	importParts    []string
	seed           func(*MockAPI) error
	computedFields map[string]map[string]interface{}
	computedChecks map[string]string
	// settle is what the API shows once the write's work is done (a status of running): the mock
	// assigns it on create, so the provider's wait for it ends.
	settle map[string]interface{}
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
		name:  "alert",
		addr:  "betterstack_alert.test",
		idKey: "id",
		createConfig: `resource "betterstack_alert" "test" {
  name         = "example"
  alert_type   = "threshold"
  dashboard_id = "12345"
  chart_id     = "12345"
}`,
		updateConfig: `resource "betterstack_alert" "test" {
  name         = "example"
  alert_type   = "threshold"
  dashboard_id = "12345"
  chart_id     = "12345"
  string_value = "example-updated"
}`,
		replaceConfig: `resource "betterstack_alert" "test" {
  name         = "example"
  alert_type   = "threshold"
  dashboard_id = "12345-updated"
  chart_id     = "12345"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "application",
		addr:  "betterstack_application.test",
		idKey: "id",
		createConfig: `resource "betterstack_application" "test" {
  name                = "example"
  platform            = "aiohttp_errors"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_application" "test" {
  name                 = "example"
  platform             = "aiohttp_errors"
  application_group_id = "12345-updated"
  deletion_protection  = false
}`,
		replaceConfig: `resource "betterstack_application" "test" {
  name                = "example"
  platform            = "android_errors"
  deletion_protection = false
}`,
		importIgnore:   []string{"deletion_protection"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "application_group",
		addr:  "betterstack_application_group.test",
		idKey: "id",
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
		name:  "aws_log_group_subscription",
		addr:  "betterstack_aws_log_group_subscription.test",
		idKey: "id",
		createConfig: `resource "betterstack_aws_log_group_subscription" "test" {
  source_id  = "12345"
  region     = "example"
  name       = "example"
  subscribed = true
}`,
		replaceConfig: `resource "betterstack_aws_log_group_subscription" "test" {
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
		name:  "catalog_record",
		addr:  "betterstack_catalog_record.test",
		idKey: "id",
		createConfig: `resource "betterstack_catalog_record" "test" {
  catalog_relation_id = "12345"
  attributes          = [{ attribute = { id = "example" }, values = [{ type = "example" }] }]
}`,
		importIgnore:   []string{},
		importParts:    []string{"catalog_relation_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "catalog_relation",
		addr:  "betterstack_catalog_relation.test",
		idKey: "id",
		createConfig: `resource "betterstack_catalog_relation" "test" {
  name       = "example"
  attributes = [{ name = "example" }]
}`,
		updateConfig: `resource "betterstack_catalog_relation" "test" {
  name        = "example"
  attributes  = [{ name = "example" }]
  description = "example-updated"
}`,
		replaceConfig: `resource "betterstack_catalog_relation" "test" {
  name            = "example"
  attributes      = [{ name = "example" }]
  organization_id = "12345-updated"
}`,
		importIgnore:   []string{"organization_id"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "collector",
		addr:  "betterstack_collector.test",
		idKey: "id",
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
		name:  "collector_target",
		addr:  "betterstack_collector_target.test",
		idKey: "id",
		createConfig: `resource "betterstack_collector_target" "test" {
  collector_id = "12345"
  kind         = "postgres"
  host         = "example"
}`,
		updateConfig: `resource "betterstack_collector_target" "test" {
  collector_id = "12345"
  kind         = "postgres"
  host         = "example-updated"
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
		name:  "connection",
		addr:  "betterstack_connection.test",
		idKey: "id",
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
		name:  "dashboard",
		addr:  "betterstack_dashboard.test",
		idKey: "id",
		createConfig: `resource "betterstack_dashboard" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_dashboard" "test" {
  name               = "example"
  dashboard_group_id = "12345-updated"
}`,
		replaceConfig: `resource "betterstack_dashboard" "test" {
  name    = "example"
  product = "rum"
}`,
		importIgnore:   []string{"resource_group_name", "source_id", "template_id"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "dashboard_chart",
		addr:  "betterstack_dashboard_chart.test",
		idKey: "id",
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
		importIgnore:   []string{},
		importParts:    []string{"dashboard_id", "id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "dashboard_group",
		addr:  "betterstack_dashboard_group.test",
		idKey: "id",
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
		name:  "dashboard_section",
		addr:  "betterstack_dashboard_section.test",
		idKey: "id",
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
		name:  "errors_jira_integration",
		addr:  "betterstack_errors_jira_integration.test",
		idKey: "id",
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
		name:  "errors_linear_integration",
		addr:  "betterstack_errors_linear_integration.test",
		idKey: "id",
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
		name:  "errors_microsoft_teams_integration",
		addr:  "betterstack_errors_microsoft_teams_integration.test",
		idKey: "id",
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
		name:  "errors_slack_integration",
		addr:  "betterstack_errors_slack_integration.test",
		idKey: "id",
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
		name:  "escalation_policy",
		addr:  "betterstack_escalation_policy.test",
		idKey: "id",
		createConfig: `resource "betterstack_escalation_policy" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_escalation_policy" "test" {
  name                       = "example"
  escalation_policy_group_id = "12345-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "escalation_policy_group",
		addr:  "betterstack_escalation_policy_group.test",
		idKey: "id",
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
		name:  "exploration",
		addr:  "betterstack_exploration.test",
		idKey: "id",
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
		replaceConfig: `resource "betterstack_exploration" "test" {
  name        = "example"
  chart       = { chart_type = "line_chart" }
  queries     = [{ query_type = "sql_expression" }]
  source_type = "logs"
}`,
		importIgnore:   []string{"source_id", "source_type"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "exploration_group",
		addr:  "betterstack_exploration_group.test",
		idKey: "id",
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
		name:  "extracted_metric",
		addr:  "betterstack_extracted_metric.test",
		idKey: "id",
		createConfig: `resource "betterstack_extracted_metric" "test" {
  source_id      = "12345"
  name           = "example"
  sql_expression = "example"
}`,
		updateConfig: `resource "betterstack_extracted_metric" "test" {
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
		name:  "heartbeat",
		addr:  "betterstack_heartbeat.test",
		idKey: "id",
		createConfig: `resource "betterstack_heartbeat" "test" {
  name                = "example"
  period              = 1
  grace               = 1
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_heartbeat" "test" {
  name                = "example"
  period              = 1
  grace               = 1
  heartbeat_group_id  = "12345-updated"
  deletion_protection = false
}`,
		importIgnore:   []string{"deletion_protection"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "heartbeat_group",
		addr:  "betterstack_heartbeat_group.test",
		idKey: "id",
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
		name:  "incidents_amazon_sns_integration",
		addr:  "betterstack_incidents_amazon_sns_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_amazon_sns_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_amazon_sns_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_aws_cloudwatch_integration",
		addr:  "betterstack_incidents_aws_cloudwatch_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_aws_cloudwatch_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_aws_cloudwatch_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_azure_integration",
		addr:  "betterstack_incidents_azure_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_azure_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_azure_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_datadog_integration",
		addr:  "betterstack_incidents_datadog_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_datadog_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_datadog_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_elastic_integration",
		addr:  "betterstack_incidents_elastic_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_elastic_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_elastic_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_google_monitoring_integration",
		addr:  "betterstack_incidents_google_monitoring_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_google_monitoring_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_google_monitoring_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_grafana_integration",
		addr:  "betterstack_incidents_grafana_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_grafana_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_grafana_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_incoming_email_integration",
		addr:  "betterstack_incidents_incoming_email_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_incoming_email_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_incoming_email_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_incoming_webhook_integration",
		addr:  "betterstack_incidents_incoming_webhook_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_incoming_webhook_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_incoming_webhook_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_jira_integration",
		addr:  "betterstack_incidents_jira_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_jira_integration" "test" {
  id = "12345"
}`,
		updateConfig: `resource "betterstack_incidents_jira_integration" "test" {
  id   = "12345"
  name = "example-updated"
}`,
		importIgnore: []string{},
		importParts:  []string{"id"},
		seed: func(mock *MockAPI) error {
			return mock.Seed("incidents_jira_integration", map[string]string{}, "12345")
		},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_new_relic_integration",
		addr:  "betterstack_incidents_new_relic_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_new_relic_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_new_relic_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_outgoing_email_integration",
		addr:  "betterstack_incidents_outgoing_email_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_outgoing_email_integration" "test" {
  recipients = [{ email = "test@example.com" }]
}`,
		updateConfig: `resource "betterstack_incidents_outgoing_email_integration" "test" {
  recipients = [{ email = "test@example.com" }]
  name       = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_outgoing_pagerduty_integration",
		addr:  "betterstack_incidents_outgoing_pagerduty_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_outgoing_pagerduty_integration" "test" {
  name     = "example"
  key      = "example"
  severity = "info"
}`,
		updateConfig: `resource "betterstack_incidents_outgoing_pagerduty_integration" "test" {
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
		name:  "incidents_outgoing_webhook_integration",
		addr:  "betterstack_incidents_outgoing_webhook_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_outgoing_webhook_integration" "test" {
  url          = "https://example.com"
  trigger_type = "incident_change"
}`,
		updateConfig: `resource "betterstack_incidents_outgoing_webhook_integration" "test" {
  url          = "https://example.com"
  trigger_type = "incident_change"
  name         = "example-updated"
}`,
		replaceConfig: `resource "betterstack_incidents_outgoing_webhook_integration" "test" {
  url          = "https://example.com"
  trigger_type = "on_call_change"
}`,
		importIgnore:   []string{"custom_webhook_template_attributes"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_prometheus_integration",
		addr:  "betterstack_incidents_prometheus_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_prometheus_integration" "test" {
}`,
		updateConfig: `resource "betterstack_incidents_prometheus_integration" "test" {
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "incidents_splunk_on_call_integration",
		addr:  "betterstack_incidents_splunk_on_call_integration.test",
		idKey: "id",
		createConfig: `resource "betterstack_incidents_splunk_on_call_integration" "test" {
  url = "https://example.com"
}`,
		updateConfig: `resource "betterstack_incidents_splunk_on_call_integration" "test" {
  url  = "https://example.com"
  name = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "metadata",
		addr:  "betterstack_metadata.test",
		idKey: "id",
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
		name:  "monitor",
		addr:  "betterstack_monitor.test",
		idKey: "id",
		createConfig: `resource "betterstack_monitor" "test" {
  url                 = "https://example.com"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_monitor" "test" {
  url                 = "https://example.com/updated"
  deletion_protection = false
}`,
		importIgnore:   []string{"auth_password", "deletion_protection"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "monitor_group",
		addr:  "betterstack_monitor_group.test",
		idKey: "id",
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
		name:  "on_call_rotation",
		addr:  "betterstack_on_call_rotation.test",
		idKey: "id",
		createConfig: `resource "betterstack_on_call_rotation" "test" {
  id                 = "example"
  users              = ["example"]
  rotation_length    = 1
  rotation_interval  = "hour"
  start_rotations_at = "2026-01-01T00:00:00Z"
}`,
		updateConfig: `resource "betterstack_on_call_rotation" "test" {
  id                 = "example"
  users              = ["example"]
  rotation_length    = 1
  rotation_interval  = "hour"
  start_rotations_at = "2026-01-01T00:00:00Z"
  timezone           = "UTC-updated"
}`,
		importIgnore:   []string{"apply_from", "apply_from_timezone"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "on_call_schedule",
		addr:  "betterstack_on_call_schedule.test",
		idKey: "id",
		createConfig: `resource "betterstack_on_call_schedule" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_on_call_schedule" "test" {
  name               = "example"
  slack_group_handle = "example-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "release",
		addr:  "betterstack_release.test",
		idKey: "id",
		createConfig: `resource "betterstack_release" "test" {
  version        = "example"
  application_id = "12345"
}`,
		replaceConfig: `resource "betterstack_release" "test" {
  version        = "example"
  application_id = "12345"
  team_id        = "12345-updated"
}`,
		importIgnore:   []string{"environment", "team_id"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "severity",
		addr:  "betterstack_severity.test",
		idKey: "id",
		createConfig: `resource "betterstack_severity" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_severity" "test" {
  name              = "example"
  severity_group_id = "12345-updated"
}`,
		importIgnore:   []string{},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "severity_group",
		addr:  "betterstack_severity_group.test",
		idKey: "id",
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
		name:  "source",
		addr:  "betterstack_source.test",
		idKey: "id",
		createConfig: `resource "betterstack_source" "test" {
  name                = "example"
  platform            = "apache2"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_source" "test" {
  name                = "example"
  platform            = "apache2"
  source_group_id     = "12345-updated"
  deletion_protection = false
}`,
		replaceConfig: `resource "betterstack_source" "test" {
  name                = "example"
  platform            = "aws"
  deletion_protection = false
}`,
		importIgnore:   []string{"deletion_protection", "scrape_request_basic_auth_password"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "source_aws_account",
		addr:  "betterstack_source_aws_account.test",
		idKey: "source_id",
		createConfig: `resource "betterstack_source_aws_account" "test" {
  source_id       = "12345"
  aws_role_arn    = "example"
  aws_external_id = "12345"
}`,
		updateConfig: `resource "betterstack_source_aws_account" "test" {
  source_id               = "12345"
  aws_role_arn            = "example"
  aws_external_id         = "12345"
  aws_auto_sub_log_groups = true
}`,
		importIgnore: []string{"aws_account_id", "aws_external_id", "aws_role_arn"},
		importParts:  []string{"source_id"},
		seed: func(mock *MockAPI) error {
			return mock.Seed("source_aws_account", map[string]string{}, "12345")
		},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "source_gcp_project",
		addr:  "betterstack_source_gcp_project.test",
		idKey: "source_id",
		createConfig: `resource "betterstack_source_gcp_project" "test" {
  source_id          = "12345"
  gcp_project_id     = "12345"
  gcp_project_number = "example"
}`,
		updateConfig: `resource "betterstack_source_gcp_project" "test" {
  source_id          = "12345"
  gcp_project_id     = "12345-updated"
  gcp_project_number = "example"
}`,
		importIgnore: []string{"gcp_account_id", "gcp_project_id", "gcp_project_number"},
		importParts:  []string{"source_id"},
		seed: func(mock *MockAPI) error {
			return mock.Seed("source_gcp_project", map[string]string{}, "12345")
		},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "source_group",
		addr:  "betterstack_source_group.test",
		idKey: "id",
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
		name:  "status_page",
		addr:  "betterstack_status_page.test",
		idKey: "id",
		createConfig: `resource "betterstack_status_page" "test" {
  company_name        = "example"
  subdomain           = "example-page"
  timezone            = "UTC"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_status_page" "test" {
  company_name        = "example"
  subdomain           = "example-page"
  timezone            = "UTC"
  company_url         = "https://example.com/updated"
  deletion_protection = false
}`,
		replaceConfig: `resource "betterstack_status_page" "test" {
  company_name        = "example"
  subdomain           = "example-page"
  timezone            = "UTC"
  organization_id     = "12345-updated"
  deletion_protection = false
}`,
		importIgnore:   []string{"deletion_protection", "password"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "status_page_group",
		addr:  "betterstack_status_page_group.test",
		idKey: "id",
		createConfig: `resource "betterstack_status_page_group" "test" {
  name = "example"
}`,
		updateConfig: `resource "betterstack_status_page_group" "test" {
  name = "example-updated"
}`,
		replaceConfig: `resource "betterstack_status_page_group" "test" {
  name            = "example"
  organization_id = "12345-updated"
}`,
		importIgnore:   []string{"organization_id"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "team",
		addr:  "betterstack_team.test",
		idKey: "id",
		createConfig: `resource "betterstack_team" "test" {
  name                = "example"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_team" "test" {
  name                = "example"
  logo_remote_url     = "https://example.com/updated"
  deletion_protection = false
}`,
		importIgnore:   []string{"deletion_protection"},
		importParts:    []string{"id"},
		computedFields: map[string]map[string]interface{}{},
		computedChecks: map[string]string{},
	},
	{
		name:  "team_member",
		addr:  "betterstack_team_member.test",
		idKey: "id",
		createConfig: `resource "betterstack_team_member" "test" {
  email               = "test@example.com"
  deletion_protection = false
}`,
		updateConfig: `resource "betterstack_team_member" "test" {
  email               = "test@example.com"
  role_id             = "12345-updated"
  deletion_protection = false
}`,
		replaceConfig: `resource "betterstack_team_member" "test" {
  email               = "test@example.com-updated"
  deletion_protection = false
}`,
		importIgnore:   []string{"deletion_protection"},
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
			if tc.settle != nil {
				mock.InjectOnCreate(tc.name, tc.settle)
			}
			mock.InjectItemFields(tc.name, tc.computedFields)
			checks := []resource.TestCheckFunc{resource.TestCheckResourceAttrSet(tc.addr, tc.idKey)}
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
			for _, config := range tc.inPlaceConfigs {
				// What the API changes in place is never a replacement.
				steps = append(steps, resource.TestStep{
					Config: providerConfig(mock.URL()) + config,
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(tc.addr, plancheck.ResourceActionUpdate)},
					},
				})
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
				ResourceName:                         tc.addr,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: tc.idKey,
				ImportStateVerifyIgnore:              tc.importIgnore,
				ImportStateIdFunc:                    importID(tc.addr, tc.importParts),
			})
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testProviderFactories,
				Steps:                    steps,
			})
		})
	}
}
