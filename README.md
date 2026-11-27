# Better Stack Terraform provider

The `betterstack` provider manages [Better Stack](https://betterstack.com) with Terraform. One provider
and one API token cover uptime monitoring, on-call, status pages, telemetry, dashboards and error tracking,
with 56 resources and 59 data sources.

- Docs: [Terraform Registry](https://registry.terraform.io/providers/BetterStackHQ/betterstack/latest/docs), with every resource's settings, an example and its import ID
- Moving from the `better-uptime` or `logtail` provider: [migration guide](https://registry.terraform.io/providers/BetterStackHQ/betterstack/latest/docs/guides/migration)

## Example usage

```terraform
terraform {
  required_providers {
    betterstack = {
      source = "BetterStackHQ/betterstack"
    }
  }
}

provider "betterstack" {
  team_id = "t123" # only for a global API token
}

resource "betterstack_monitor" "example" {
  url = "https://uptime.betterstack.com"
}
```

Set your API token in the `BETTERSTACK_API_TOKEN` environment variable, or in the provider's `api_token`, then run
`terraform apply`. Use a global API token from [Better Stack → API tokens → Global tokens](https://betterstack.com/settings/global-api-tokens),
or a team API token from [Better Stack → API tokens → Team-based tokens](https://betterstack.com/settings/api-tokens/0).
`team_id` picks the team when the token can reach several, as a global API token can. Write it as the team's
ID from Better Stack URLs, such as `t123`.

Monitors, heartbeats, sources and some other resources start with `deletion_protection = true`, so
`terraform destroy` stops at them until you set it to `false`.

## Resources

**Incidents** (31): `betterstack_catalog_record`, `betterstack_catalog_relation`, `betterstack_escalation_policy`, `betterstack_escalation_policy_group`, `betterstack_heartbeat`, `betterstack_heartbeat_group`, `betterstack_incidents_amazon_sns_integration`, `betterstack_incidents_aws_cloudwatch_integration`, `betterstack_incidents_azure_integration`, `betterstack_incidents_datadog_integration`, `betterstack_incidents_elastic_integration`, `betterstack_incidents_google_monitoring_integration`, `betterstack_incidents_grafana_integration`, `betterstack_incidents_incoming_email_integration`, `betterstack_incidents_incoming_webhook_integration`, `betterstack_incidents_jira_integration`, `betterstack_incidents_new_relic_integration`, `betterstack_incidents_outgoing_email_integration`, `betterstack_incidents_outgoing_pagerduty_integration`, `betterstack_incidents_outgoing_webhook_integration`, `betterstack_incidents_prometheus_integration`, `betterstack_incidents_splunk_on_call_integration`, `betterstack_metadata`, `betterstack_monitor`, `betterstack_monitor_group`, `betterstack_on_call_rotation`, `betterstack_on_call_schedule`, `betterstack_severity`, `betterstack_severity_group`, `betterstack_status_page`, `betterstack_status_page_group`

**Telemetry** (16): `betterstack_alert`, `betterstack_aws_log_group_subscription`, `betterstack_collector`, `betterstack_collector_target`, `betterstack_connection`, `betterstack_dashboard`, `betterstack_dashboard_chart`, `betterstack_dashboard_group`, `betterstack_dashboard_section`, `betterstack_exploration`, `betterstack_exploration_group`, `betterstack_extracted_metric`, `betterstack_source`, `betterstack_source_aws_account`, `betterstack_source_gcp_project`, `betterstack_source_group`

**Errors** (7): `betterstack_application`, `betterstack_application_group`, `betterstack_errors_jira_integration`, `betterstack_errors_linear_integration`, `betterstack_errors_microsoft_teams_integration`, `betterstack_errors_slack_integration`, `betterstack_release`

**Account** (2): `betterstack_team`, `betterstack_team_member`

## About this repository

This repository is published from Better Stack's API description, where the provider is built and tested:
each release replaces its contents. Please report bugs and requests in this repository's issues.

Licensed under Apache-2.0 (`LICENSE`).
