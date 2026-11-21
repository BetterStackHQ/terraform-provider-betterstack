---
page_title: "betterstack_heartbeat Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A heartbeat that expects a periodic ping and alerts if one fails to arrive in time.
---

# betterstack_heartbeat (Data Source)

A heartbeat that expects a periodic ping and alerts if one fails to arrive in time.

## Example Usage

```terraform
data "betterstack_heartbeat" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the heartbeat to look up.

### Read-Only

- `team_name` (String) Name of the team that owns this heartbeat.
- `metadata` (Attributes Map, see [nested schema](#nestedatt--metadata)) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. Team references are returned as team IDs such as `t123`. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`.
- `name` (String) Heartbeat name.
- `url` (String, Sensitive) URL to send pings to.
- `status` (String) Current heartbeat status.
- `paused` (Boolean) Whether the heartbeat is paused.
- `period` (Number) Expected time between pings, in seconds. At least 30 and less than a year.
- `grace` (Number) How long to wait after the period before alerting, in seconds. At least 0.
- `heartbeat_group_id` (String) ID of the heartbeat group this heartbeat belongs to.
- `sort_index` (Number) Display order within the group.
- `email` (Boolean) Alert by e-mail.
- `sms` (Boolean) Alert by SMS.
- `call` (Boolean) Alert by phone call.
- `push` (Boolean) Alert by push notification.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode.
- `team_wait` (Number) Seconds before escalating to the whole team.
- `escalation_policy_id` (String) ID of the escalation policy used for alerts. Returned as a string. You can send it as a number or a string.
- `server_timezone` (String) IANA time zone, such as `Europe/London`, that keeps expected pings at the same local time of day. Only accepted for a period of 1 hour (3600 seconds) or more.
- `maintenance_from` (String) Start time of the maintenance window, as `HH:MM:SS`.
- `maintenance_to` (String) End time of the maintenance window, as `HH:MM:SS`.
- `maintenance_timezone` (String) Time zone of the maintenance window.
- `maintenance_days` (Set of String) Days of the week the maintenance window applies to.
- `paused_at` (String (RFC 3339)) When the heartbeat was paused.
- `created_at` (String (RFC 3339)) When the heartbeat was created.
- `updated_at` (String (RFC 3339)) When the heartbeat was last updated.

<a id="nestedatt--metadata"></a>
### Nested Schema for `metadata`

- `values` (Attributes List, see [nested schema](#nestedatt--metadata--values))

<a id="nestedatt--metadata--values"></a>
### Nested Schema for `metadata.values`

- `type` (String) `String`, or the type of the referenced item.
- `value` (String) The text of a String value.
- `item_id` (String) ID of the referenced item: a number, or for a team its ID starting with `t`, such as `t123`.
- `name` (String) Referenced item's name.
- `email` (String) Referenced user's e-mail.

## Endpoints

- `GET /api/v3/heartbeats`
- `GET /api/v3/heartbeats/:id`
- `POST /api/v3/heartbeats`
- `PATCH /api/v3/heartbeats/:id`
- `DELETE /api/v3/heartbeats/:id`
