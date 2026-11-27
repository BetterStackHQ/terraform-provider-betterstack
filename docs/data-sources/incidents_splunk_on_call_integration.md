---
page_title: "betterstack_incidents_splunk_on_call_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Forwards Better Stack incidents to Splunk On-Call, formerly VictorOps.
---

# betterstack_incidents_splunk_on_call_integration (Data Source)

Forwards Better Stack incidents to Splunk On-Call, formerly VictorOps.

## Example Usage

```terraform
data "betterstack_incidents_splunk_on_call_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the Splunk On-Call integration to look up.

### Optional

- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) Integration name.
- `url` (String, Sensitive) Splunk On-Call REST endpoint URL.
- `notify_alongside_primary_responder` (Boolean) Whether to alert this Splunk On-Call integration alongside the primary responder for incidents with no escalation policy. For an incident with an escalation policy, it's alerted only when a step of that policy notifies it.
- `paused` (Boolean) Whether the integration is paused. A paused integration sends nothing to Splunk On-Call.
- `team_name` (String) Name of the team that owns this integration.
