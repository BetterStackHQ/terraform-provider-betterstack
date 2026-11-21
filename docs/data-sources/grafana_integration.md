---
page_title: "betterstack_grafana_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives alerts from Grafana and opens Better Stack incidents.
---

# betterstack_grafana_integration (Data Source)

Receives alerts from Grafana and opens Better Stack incidents.

## Example Usage

```terraform
data "betterstack_grafana_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the grafana_integration to look up.

### Read-Only

- `name` (String) Integration name.
- `team_name` (String) Name of the team that owns this integration.
- `escalation_policy_id` (String) ID of the escalation policy used for alerts.
- `webhook_url` (String, Sensitive) The URL to paste into Grafana. Alerts Grafana sends to it open incidents.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents.
- `call` (Boolean) Alert by phone call.
- `sms` (Boolean) Alert by SMS.
- `email` (Boolean) Alert by e-mail.
- `push` (Boolean) Alert by push notification.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode.
- `team_wait` (Number) Seconds before escalating to the whole team.
- `recovery_period` (Number) Seconds a resolved alert must hold before the incident resolves. Has no effect for this integration, since an incident resolves as soon as Grafana reports the alert as resolved.

## Endpoints

- `GET /api/v3/grafana-integrations`
- `GET /api/v3/grafana-integrations/:id`
- `POST /api/v3/grafana-integrations`
- `PATCH /api/v3/grafana-integrations/:id`
- `DELETE /api/v3/grafana-integrations/:id`
