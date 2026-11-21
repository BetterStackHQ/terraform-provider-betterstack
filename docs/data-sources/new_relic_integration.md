---
page_title: "betterstack_new_relic_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives alerts from New Relic and opens Better Stack incidents.
---

# betterstack_new_relic_integration (Data Source)

Receives alerts from New Relic and opens Better Stack incidents.

## Example Usage

```terraform
data "betterstack_new_relic_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the new_relic_integration to look up.

### Read-Only

- `name` (String) Integration name.
- `team_name` (String) Name of the team that owns this integration.
- `escalation_policy_id` (String) ID of the escalation policy used for alerts.
- `webhook_url` (String, Sensitive) The URL to paste into New Relic. Alerts New Relic sends to it open incidents.
- `alerting_rule` (String) Which New Relic alert levels open incidents: alert only, or alert and warn. Valid values: `alert`, `alert_and_warn`.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents.
- `call` (Boolean) Alert by phone call.
- `sms` (Boolean) Alert by SMS.
- `email` (Boolean) Alert by e-mail.
- `push` (Boolean) Alert by push notification.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode.
- `team_wait` (Number) Seconds before escalating to the whole team.
- `recovery_period` (Number) Seconds a resolved alert must hold before the incident resolves. Has no effect for this integration, since an incident resolves as soon as New Relic reports the alert as resolved.

## Endpoints

- `GET /api/v3/new-relic-integrations`
- `GET /api/v3/new-relic-integrations/:id`
- `POST /api/v3/new-relic-integrations`
- `PATCH /api/v3/new-relic-integrations/:id`
- `DELETE /api/v3/new-relic-integrations/:id`
