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
  name = "Testing heartbeat"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Heartbeat name.
- `status` (String) The heartbeat's current status.
  - `paused`: The heartbeat is paused.
  - `pending`: The first request hasn't arrived yet.
  - `maintenance`: The heartbeat is in its maintenance window.
  - `up`: There's no open incident.
  - `validating`: Requests arrive again, and the open incident resolves once the recovery period passes.
  - `down`: A request didn't arrive in time or reported a failure.

### Optional

- `team_id` (String) Team that owns this heartbeat, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `url` (String, Sensitive) The unique URL your service requests to report a heartbeat.
- `paused` (Boolean) Whether the heartbeat is paused. A paused heartbeat doesn't notify you about downtime.
- `period` (Number) How often the heartbeat expects a request, in seconds. At least `30` and less than a year.
- `grace` (Number) How late a request can arrive after `period`, in seconds, before it counts as missed. At least `0` and less than a year.
- `heartbeat_group_id` (String) ID of the heartbeat group this heartbeat belongs to.
- `sort_index` (Number) A sort value you can set through the API. The dashboard doesn't use it to order heartbeats.
- `email` (Boolean) Whether to send an e-mail to the on-call person. Ignored while an escalation policy is set.
- `sms` (Boolean) Whether to send an SMS to the on-call person. Ignored while an escalation policy is set.
- `call` (Boolean) Whether to call the on-call person. Ignored while an escalation policy is set.
- `push` (Boolean) Whether to send a push notification to the on-call person. Ignored while an escalation policy is set.
- `critical_alert` (Boolean) Whether to send a critical push notification that ignores the mute switch and Do Not Disturb mode. Ignored while an escalation policy is set.
- `team_wait` (Number) Seconds to wait before escalating to the whole team. `0` alerts the whole team right away, and `null` never escalates to the team. Ignored while an escalation policy is set.
- `escalation_policy_id` (String) ID of the escalation policy for alerts, or `null` to alert the current on-call person directly. Get the ID via [List escalation policies](https://betterstack.com/docs/uptime/api/list-all-escalation-policies/).
- `server_timezone` (String) The server's time zone, an IANA name like `Europe/Berlin`, so missed heartbeats are counted in its local time across daylight saving changes.
- `maintenance_from` (String) Start time of the daily maintenance window, like `01:00`. We don't check the heartbeat during the window. Returned as `HH:MM:SS`, or `null` when no window is set.
- `maintenance_to` (String) End time of the daily maintenance window, like `03:00`. Required with `maintenance_from`. A window that ends before it starts runs overnight. Returned as `HH:MM:SS`, or `null` when no window is set.
- `maintenance_timezone` (String) Time zone of the maintenance window, as a Rails time zone name like `Amsterdam` or an IANA name like `Europe/Amsterdam`. Returned as the Rails name. `null` means UTC.
- `maintenance_days` (Set of String) Days of the week the maintenance window applies to, any of `mon`, `tue`, `wed`, `thu`, `fri`, `sat` and `sun`. An overnight window counts for the day it starts.
- `paused_at` (String (RFC 3339)) When the heartbeat was paused.
- `created_at` (String (RFC 3339)) When the heartbeat was created.
- `updated_at` (String (RFC 3339)) When the heartbeat was last updated.
- `team_name` (String) Name of the team that owns this heartbeat.
- `metadata` (Attributes Map) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. User references are returned as `u123` and team references as `t123`. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). Values under one key must all be the same type and can't repeat. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. (see [below for nested schema](#nestedatt--metadata))
- `deletion_protection` (Boolean) Whether destroying or replacing this heartbeat is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.

<a id="nestedatt--metadata"></a>
### Nested Schema for `metadata`

Read-Only:

- `values` (Attributes List) (see [below for nested schema](#nestedatt--metadata--values))

<a id="nestedatt--metadata--values"></a>
### Nested Schema for `metadata.values`

Read-Only:

- `type` (String) `String`, or the type of the referenced item, which is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`.
- `value` (String) The text of a `String` value.
- `item_id` (String) ID of the referenced item. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `name` (String) Referenced item's name, or `null` when the item has no name or no longer exists. An item can also be found by its name instead of `item_id`.
- `email` (String) Referenced user's e-mail, returned for a User. A user can also be found by e-mail.
