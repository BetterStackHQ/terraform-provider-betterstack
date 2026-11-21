---
page_title: "betterstack_azure_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives alerts from Azure and opens Better Stack incidents.
---

# betterstack_azure_integration (Data Source)

Receives alerts from Azure and opens Better Stack incidents.

## Example Usage

```terraform
data "betterstack_azure_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the azure_integration to look up.

### Read-Only

- `name` (String) Integration name.
- `team_name` (String) Name of the team that owns this integration.
- `escalation_policy_id` (String) ID of the escalation policy used for alerts.
- `webhook_url` (String, Sensitive) The URL to paste into Azure. Alerts Azure sends to it open incidents.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents.
- `call` (Boolean) Alert by phone call.
- `sms` (Boolean) Alert by SMS.
- `email` (Boolean) Alert by e-mail.
- `push` (Boolean) Alert by push notification.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode.
- `team_wait` (Number) Seconds before escalating to the whole team.
- `recovery_period` (Number) Seconds a resolved alert must hold before the incident resolves. Has no effect for this integration, since an incident resolves as soon as Azure reports the alert as resolved.

## Endpoints

- `GET /api/v3/azure-integrations`
- `GET /api/v3/azure-integrations/:id`
- `POST /api/v3/azure-integrations`
- `PATCH /api/v3/azure-integrations/:id`
- `DELETE /api/v3/azure-integrations/:id`
