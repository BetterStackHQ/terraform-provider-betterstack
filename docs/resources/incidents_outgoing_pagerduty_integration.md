---
page_title: "betterstack_incidents_outgoing_pagerduty_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Forwards Better Stack incidents to a PagerDuty service.
---

# betterstack_incidents_outgoing_pagerduty_integration (Resource)

Forwards Better Stack incidents to a PagerDuty service.

-> Replaces `betteruptime_pagerduty_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Send incidents to a PagerDuty service as critical

```terraform
resource "betterstack_incidents_outgoing_pagerduty_integration" "send_incidents_to_a_pagerduty_service_as_critical" {
  name     = "Production on-call"
  key      = "e93facc04764012d7bfb002500d5d1a6"
  severity = "critical"
}
```

### Send warnings only when an escalation policy step notifies PagerDuty

```terraform
resource "betterstack_incidents_outgoing_pagerduty_integration" "send_warnings_only_when_an_escalation_policy_step_notifies_pagerduty" {
  name                               = "Staging alerts"
  key                                = "a1b2c3d4e5f60718293a4b5c6d7e8f90"
  severity                           = "warning"
  notify_alongside_primary_responder = false
}
```

## Schema

### Required

- `name` (String) Integration name.
- `key` (String, Sensitive) PagerDuty integration key, also called the routing key.
- `severity` (String) PagerDuty severity sent with the event. Valid values: `info`, `warning`, `error`, `critical`.

### Optional

- `notify_alongside_primary_responder` (Boolean) Whether to alert this PagerDuty integration alongside the primary responder for incidents with no escalation policy. For an incident with an escalation policy, it's alerted only when a step of that policy notifies it. Defaults to `true`.
- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this outgoing PagerDuty integration.
- `team_name` (String) Name of the team that owns this integration.

## Import

Import an existing outgoing PagerDuty integration by its ID. Get the ID via [List PagerDuty integrations](https://betterstack.com/docs/uptime/api/list-all-pagerduty-integrations/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_incidents_outgoing_pagerduty_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_incidents_outgoing_pagerduty_integration.example 12345
```
