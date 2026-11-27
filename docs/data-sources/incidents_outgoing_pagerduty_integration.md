---
page_title: "betterstack_incidents_outgoing_pagerduty_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Forwards Better Stack incidents to a PagerDuty service.
---

# betterstack_incidents_outgoing_pagerduty_integration (Data Source)

Forwards Better Stack incidents to a PagerDuty service.

## Example Usage

```terraform
data "betterstack_incidents_outgoing_pagerduty_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the outgoing PagerDuty integration to look up.

### Optional

- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) Integration name.
- `key` (String, Sensitive) PagerDuty integration key, also called the routing key.
- `severity` (String) PagerDuty severity sent with the event. Valid values: `info`, `warning`, `error`, `critical`.
- `notify_alongside_primary_responder` (Boolean) Whether to alert this PagerDuty integration alongside the primary responder for incidents with no escalation policy. For an incident with an escalation policy, it's alerted only when a step of that policy notifies it.
- `team_name` (String) Name of the team that owns this integration.
