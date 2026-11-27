---
page_title: "betterstack_incidents_grafana_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives alerts from Grafana and opens Better Stack incidents.
---

# betterstack_incidents_grafana_integration (Resource)

Receives alerts from Grafana and opens Better Stack incidents.

-> Replaces `betteruptime_grafana_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a Grafana integration

```terraform
resource "betterstack_incidents_grafana_integration" "example" {
  name      = "Grafana alerts"
  sms       = true
  team_wait = 180
}
```

### In an escalation policy

```terraform
resource "betterstack_escalation_policy" "escalation_policy" {
  name         = "Standard Escalation Policy"
  repeat_count = 3
  repeat_delay = 1800
  steps        = [{ type = "escalation", wait_before = 0, severity_id = "$SEVERITY_ID", step_members = [{ type = "current_on_call" }] }, { type = "escalation", wait_before = 600, severity_id = "$SEVERITY_ID", step_members = [{ type = "entire_team" }] }]
}

resource "betterstack_incidents_grafana_integration" "in_escalation_policy" {
  name                 = "Grafana alerts"
  sms                  = true
  team_wait            = 180
  escalation_policy_id = betterstack_escalation_policy.escalation_policy.id
}
```

## Schema

### Optional

- `name` (String) Integration name.
- `escalation_policy_id` (String) ID of the escalation policy for alerts, or `null` to alert the current on-call person directly.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents. Defaults to `false`.
- `call` (Boolean) Whether to call the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `sms` (Boolean) Whether to send an SMS to the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `email` (Boolean) Whether to send an e-mail to the on-call person. Ignored while an escalation policy is set. Defaults to `true`.
- `push` (Boolean) Whether to send a push notification to the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `critical_alert` (Boolean) Whether to send a critical push notification that ignores the mute switch and Do Not Disturb mode. Ignored while an escalation policy is set. Defaults to `false`.
- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `team_wait` (Number) Seconds to wait before escalating to the whole team. `0` alerts the whole team right away, and `null` never escalates to the team. Ignored while an escalation policy is set.

### Read-Only

- `id` (String) The ID of this Grafana integration.
- `webhook_url` (String, Sensitive) The URL to paste into Grafana. Alerts Grafana sends to it open incidents. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `team_name` (String) Name of the team that owns this integration.

-> The `alerting_rule` attribute is accepted on write but not read back.

## Import

Import an existing Grafana integration by its ID. Get the ID via [List Grafana integrations](https://betterstack.com/docs/uptime/api/list-all-grafana-integrations/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_incidents_grafana_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_incidents_grafana_integration.example 12345
```
