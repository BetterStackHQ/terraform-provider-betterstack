---
page_title: "betterstack_incidents_datadog_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives alerts from Datadog and opens Better Stack incidents.
---

# betterstack_incidents_datadog_integration (Data Source)

Receives alerts from Datadog and opens Better Stack incidents.

## Example Usage

```terraform
data "betterstack_incidents_datadog_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the Datadog integration to look up.

### Optional

- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) Integration name.
- `escalation_policy_id` (String) ID of the escalation policy for alerts, or `null` to alert the current on-call person directly.
- `webhook_url` (String, Sensitive) The URL to paste into Datadog. Alerts Datadog sends to it open incidents.
- `alerting_rule` (String) Which Datadog alert levels open incidents.
  - `alert`: Opens an incident for alerts only.
  - `alert_and_warn`: Opens an incident for alerts and warnings.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents.
- `call` (Boolean) Whether to call the on-call person. Ignored while an escalation policy is set.
- `sms` (Boolean) Whether to send an SMS to the on-call person. Ignored while an escalation policy is set.
- `email` (Boolean) Whether to send an e-mail to the on-call person. Ignored while an escalation policy is set.
- `push` (Boolean) Whether to send a push notification to the on-call person. Ignored while an escalation policy is set.
- `critical_alert` (Boolean) Whether to send a critical push notification that ignores the mute switch and Do Not Disturb mode. Ignored while an escalation policy is set.
- `team_wait` (Number) Seconds to wait before escalating to the whole team. `0` alerts the whole team right away, and `null` never escalates to the team. Ignored while an escalation policy is set.
- `team_name` (String) Name of the team that owns this integration.
