---
page_title: "betterstack_heartbeat Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A heartbeat that expects a periodic ping and alerts if one fails to arrive in time.
---

# betterstack_heartbeat (Resource)

A heartbeat that expects a periodic ping and alerts if one fails to arrive in time.

-> Replaces `betteruptime_heartbeat`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_heartbeat" "example" {
  name   = "Testing heartbeat"
  period = 10800
  grace  = 300
}
```

## Schema

### Required

- `name` (String) Heartbeat name.
- `period` (Number) Expected time between pings, in seconds. At least 30 and less than a year.
- `grace` (Number) How long to wait after the period before alerting, in seconds. At least 0.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `metadata` (Attributes Map, see [nested schema](#nestedatt--metadata)) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. Team references are returned as team IDs such as `t123`. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `paused` (Boolean) Whether the heartbeat is paused. Defaults to `false`.
- `heartbeat_group_id` (String) ID of the heartbeat group this heartbeat belongs to. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `sort_index` (Number) Display order within the group. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `email` (Boolean) Alert by e-mail. Defaults to `true`.
- `sms` (Boolean) Alert by SMS. Defaults to `false`.
- `call` (Boolean) Alert by phone call. Defaults to `false`.
- `push` (Boolean) Alert by push notification. Defaults to `false`.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode. Defaults to `false`.
- `team_wait` (Number) Seconds before escalating to the whole team. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `escalation_policy_id` (String) ID of the escalation policy used for alerts. Returned as a string. You can send it as a number or a string. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `server_timezone` (String) IANA time zone, such as `Europe/London`, that keeps expected pings at the same local time of day. Only accepted for a period of 1 hour (3600 seconds) or more. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `maintenance_from` (String) Start time of the maintenance window, as `HH:MM:SS`. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `maintenance_to` (String) End time of the maintenance window, as `HH:MM:SS`. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `maintenance_timezone` (String) Time zone of the maintenance window. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `maintenance_days` (Set of String) Days of the week the maintenance window applies to. Defaults to `["mon", "tue", "wed", "thu", "fri", "sat", "sun"]`.

### Read-Only

- `id` (String) The id of this heartbeat.
- `team_name` (String) Name of the team that owns this heartbeat.
- `url` (String, Sensitive) URL to send pings to. Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `status` (String) Current heartbeat status.
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

-> A metadata object you send replaces all keys, including keys written through the metadata endpoints (`/api/v4/metadata`). Leave metadata out to keep it unchanged, or send an empty object (`{}`) to clear it. Manage an object's metadata either here or through those endpoints, not both.

## Import

```shell
terraform import betterstack_heartbeat.example 12345
```
