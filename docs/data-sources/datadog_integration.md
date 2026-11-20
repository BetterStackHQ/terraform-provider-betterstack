---
page_title: "betterstack_datadog_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives alerts from Datadog and opens Better Stack incidents.
---

# betterstack_datadog_integration (Data Source)

Receives alerts from Datadog and opens Better Stack incidents.

## Example Usage

```terraform
data "betterstack_datadog_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the datadog_integration to look up.

### Read-Only

- `name` (String) Integration name.
- `team_name` (String) Name of the team that owns this integration.
- `policy_id` (Number) ID of the escalation policy used for alerts.
- `webhook_url` (String, Sensitive) The URL to paste into Datadog. Alerts Datadog sends to it open incidents.
- `alerting_rule` (String) Which Datadog alert levels open incidents: alert only, or alert and warn. Valid values: `alert`, `alert_and_warn`.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents.
- `call` (Boolean) Alert by phone call.
- `sms` (Boolean) Alert by SMS.
- `email` (Boolean) Alert by e-mail.
- `push` (Boolean) Alert by push notification.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode.
- `team_wait` (Number) Seconds before escalating to the whole team.
- `recovery_period` (Number) Seconds a resolved alert must hold before the incident resolves.

## Endpoints

- `GET /api/v2/datadog-integrations`
- `GET /api/v2/datadog-integrations/:id`
- `POST /api/v2/datadog-integrations`
- `PATCH /api/v2/datadog-integrations/:id`
- `DELETE /api/v2/datadog-integrations/:id`
