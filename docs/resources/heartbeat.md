---
page_title: "betterstack_heartbeat Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A heartbeat that expects a periodic ping and alerts if one fails to arrive in time.
---

# betterstack_heartbeat (Resource)

A heartbeat that expects a periodic ping and alerts if one fails to arrive in time.

~> Destroying a heartbeat deletes its data for good, so `deletion_protection` is on by default: Terraform refuses to destroy or replace it until you set `deletion_protection = false` and apply.

-> Replaces `betteruptime_heartbeat`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a heartbeat

```terraform
resource "betterstack_heartbeat" "example" {
  name   = "Testing heartbeat"
  period = 10800
  grace  = 300
}
```

### With metadata

```terraform
resource "betterstack_heartbeat" "with_metadata" {
  name     = "Testing heartbeat"
  period   = 10800
  grace    = 300
  metadata = { Environment = { values = [{ value = "production" }] }, Owner = { values = [{ type = "Team", name = "Platform" }] } }
}
```

### In a heartbeat group

```terraform
resource "betterstack_heartbeat_group" "example" {
  name = "Backend services"
}

resource "betterstack_heartbeat" "in_heartbeat_group" {
  name               = "Testing heartbeat"
  period             = 10800
  grace              = 300
  heartbeat_group_id = betterstack_heartbeat_group.example.id
}
```

## Schema

### Required

- `name` (String) Heartbeat name.
- `period` (Number) How often the heartbeat expects a request, in seconds. At least `30` and less than a year.
- `grace` (Number) How late a request can arrive after `period`, in seconds, before it counts as missed. At least `0` and less than a year.

### Optional

- `paused` (Boolean) Whether the heartbeat is paused. A paused heartbeat doesn't notify you about downtime. Set it to `false` to resume it. Defaults to `false`.
- `heartbeat_group_id` (String) ID of the heartbeat group this heartbeat belongs to.
- `sort_index` (Number) A sort value you can set through the API. The dashboard doesn't use it to order heartbeats.
- `email` (Boolean) Whether to send an e-mail to the on-call person. Ignored while an escalation policy is set. Defaults to `true`.
- `sms` (Boolean) Whether to send an SMS to the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `call` (Boolean) Whether to call the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `push` (Boolean) Whether to send a push notification to the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `critical_alert` (Boolean) Whether to send a critical push notification that ignores the mute switch and Do Not Disturb mode. Ignored while an escalation policy is set. Defaults to `false`.
- `team_id` (String) Team that owns this heartbeat, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `team_wait` (Number) Seconds to wait before escalating to the whole team. `0` alerts the whole team right away, and `null` never escalates to the team. Ignored while an escalation policy is set.
- `escalation_policy_id` (String) ID of the escalation policy for alerts, or `null` to alert the current on-call person directly. Get the ID via [List escalation policies](https://betterstack.com/docs/uptime/api/list-all-escalation-policies/).
- `server_timezone` (String) The server's time zone, an IANA name like `Europe/Berlin`, so missed heartbeats are counted in its local time across daylight saving changes. Only accepted for a `period` of at least 3600 seconds, which is 1 hour. Remove the argument or set it to an empty string to clear it.
- `maintenance_from` (String) Start time of the daily maintenance window, like `01:00`. We don't check the heartbeat during the window. Returned as `HH:MM:SS`, or `null` when no window is set.
- `maintenance_to` (String) End time of the daily maintenance window, like `03:00`. Required with `maintenance_from`. A window that ends before it starts runs overnight. Returned as `HH:MM:SS`, or `null` when no window is set.
- `maintenance_timezone` (String) Time zone of the maintenance window, as a Rails time zone name like `Amsterdam` or an IANA name like `Europe/Amsterdam`. Returned as the Rails name. `null` means UTC.
- `maintenance_days` (Set of String) Days of the week the maintenance window applies to, any of `mon`, `tue`, `wed`, `thu`, `fri`, `sat` and `sun`. An overnight window counts for the day it starts. Defaults to `["mon", "tue", "wed", "thu", "fri", "sat", "sun"]`.
- `metadata` (Attributes Map, Computed) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. User references are returned as `u123` and team references as `t123`. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). Values under one key must all be the same type and can't repeat. An update replaces all of the heartbeat's metadata with the object. `{}` removes it all. Leave `metadata` out to keep it. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. (see [below for nested schema](#nestedatt--metadata))
- `deletion_protection` (Boolean) Whether destroying or replacing this heartbeat is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.

### Read-Only

- `id` (String) The ID of this heartbeat.
- `url` (String, Sensitive) The unique URL your service requests to report a heartbeat. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `status` (String) The heartbeat's current status.
  - `paused`: The heartbeat is paused.
  - `pending`: The first request hasn't arrived yet.
  - `maintenance`: The heartbeat is in its maintenance window.
  - `up`: There's no open incident.
  - `validating`: Requests arrive again, and the open incident resolves once the recovery period passes.
  - `down`: A request didn't arrive in time or reported a failure.
- `paused_at` (String (RFC 3339)) When the heartbeat was paused.
- `created_at` (String (RFC 3339)) When the heartbeat was created.
- `updated_at` (String (RFC 3339)) When the heartbeat was last updated.
- `team_name` (String) Name of the team that owns this heartbeat.

<a id="nestedatt--metadata"></a>
### Nested Schema for `metadata`

Optional:

- `values` (Attributes List) (see [below for nested schema](#nestedatt--metadata--values))

<a id="nestedatt--metadata--values"></a>
### Nested Schema for `metadata.values`

Optional:

- `type` (String) `String`, or the type of the referenced item, which is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`.
- `value` (String) The text of a `String` value.
- `item_id` (String) ID of the referenced item. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `name` (String) Referenced item's name, or `null` when the item has no name or no longer exists. An item can also be found by its name instead of `item_id`.
- `email` (String) Referenced user's e-mail, returned for a User. A user can also be found by e-mail.

-> A metadata object replaces all keys, including keys written through the metadata endpoints at `/api/v4/metadata`. Leaving metadata out keeps it unchanged, and an empty object, `{}`, clears it. Write an object's metadata either here or through those endpoints, not both.

## Import

Import an existing heartbeat by its ID. Get the ID from the [`betterstack_heartbeat` data source](../data-sources/heartbeat.md), which finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_heartbeat.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_heartbeat.example 12345
```
