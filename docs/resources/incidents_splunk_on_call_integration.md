---
page_title: "betterstack_incidents_splunk_on_call_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Forwards Better Stack incidents to Splunk On-Call, formerly VictorOps.
---

# betterstack_incidents_splunk_on_call_integration (Resource)

Forwards Better Stack incidents to Splunk On-Call, formerly VictorOps.

-> Replaces `betteruptime_splunk_oncall_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Forward incidents to a Splunk On-Call routing key

```terraform
resource "betterstack_incidents_splunk_on_call_integration" "forward_incidents_to_a_splunk_on_call_routing_key" {
  name = "Production alerts"
  url  = "https://alert.victorops.com/integrations/generic/20131114/alert/3f1e6a2b-8c4d-4e5f-9a0b-1c2d3e4f5a6b/production"
}
```

### Forward only when an escalation policy step notifies Splunk On-Call

```terraform
resource "betterstack_incidents_splunk_on_call_integration" "forward_only_when_an_escalation_policy_step_notifies_splunk_on_call" {
  name                               = "Database escalations"
  url                                = "https://alert.victorops.com/integrations/generic/20131114/alert/3f1e6a2b-8c4d-4e5f-9a0b-1c2d3e4f5a6b/database"
  notify_alongside_primary_responder = false
}
```

## Schema

### Required

- `url` (String, Sensitive) Splunk On-Call REST endpoint URL.

### Optional

- `name` (String) Integration name.
- `notify_alongside_primary_responder` (Boolean) Whether to alert this Splunk On-Call integration alongside the primary responder for incidents with no escalation policy. For an incident with an escalation policy, it's alerted only when a step of that policy notifies it. Defaults to `true`.
- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this Splunk On-Call integration.
- `paused` (Boolean) Whether the integration is paused. A paused integration sends nothing to Splunk On-Call.
- `team_name` (String) Name of the team that owns this integration.

-> Splunk On-Call integrations and outgoing webhooks share one set of ids. The list returns only Splunk On-Call integrations, but get, update and delete also find an outgoing webhook by its id, so pass only an id this list returned.

## Import

Import an existing Splunk On-Call integration by its ID. Get the ID via [List Splunk On-Call integrations](https://betterstack.com/docs/uptime/api/list-all-splunk-on-call-integrations/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_incidents_splunk_on_call_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_incidents_splunk_on_call_integration.example 12345
```
