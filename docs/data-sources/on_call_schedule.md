---
page_title: "betterstack_on_call_schedule Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An on-call schedule the team sets up, made of a rotation and its shifts and overrides.
---

# betterstack_on_call_schedule (Data Source)

An on-call schedule the team sets up, made of a rotation and its shifts and overrides.

## Example Usage

```terraform
data "betterstack_on_call_schedule" "example" {
  name = "Primary on-call"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Schedule name. It's `null` only on a primary schedule without a name of its own.
Omitting all lookup inputs selects the API's `default` object.

### Optional

- `team_id` (String) Team that owns the schedule, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `default_schedule` (Boolean) Whether this is the team's default on-call schedule.
- `slack_group_handle` (String) The handle of the Slack user group that follows who is on call, without the @, or `null` when there's none. A group Better Stack creates can have up to 80 characters. Slack can still refuse a handle after the save, which shows in `slack_groups`.
- `use_existing_slack_group` (Boolean) Whether `slack_group_handle` names a user group that already existed in the connected Slack workspace, rather than one Better Stack created. An existing group keeps its other members. A handle one of your Slack user groups already uses is refused unless this is `true`. A handle no user group uses is refused when it's `true`. Sent without `slack_group_handle`, it applies to the current handle.
- `slack_groups` (Attributes List) The state of the handle's user group in each connected Slack workspace. (see [below for nested schema](#nestedatt--slack_groups))
- `on_call_users` (Attributes List) Users on call now, or at `from` when you get or list schedules with it. Each user's entry under relationships also has `meta` with `email`, `on_holiday` and `holiday_ends_at`. (see [below for nested schema](#nestedatt--on_call_users))
- `shifts` (Attributes List) Who is on call from `from` to `to`, oldest first. Each entry is a part of a shift with its people, and an override replaces the part of a shift it covers. Without `from` and `to`, it holds the shift running now. Returned when you get or list schedules. (see [below for nested schema](#nestedatt--shifts))
- `team_name` (String) Name of the team that owns the schedule.

<a id="nestedatt--slack_groups"></a>
### Nested Schema for `slack_groups`

Read-Only:

- `slack_workspace_id` (String) ID of the Slack workspace.
- `slack_workspace_name` (String) Name of the Slack workspace.
- `status` (String) The group's state in the workspace. When the group doesn't follow who's on call, `error` says why.
  - `pending`: Better Stack is setting the group up.
  - `active`: The group follows who's on call.
  - `permission_denied`: The Slack workspace lets only admins manage user groups.
  - `plan_required`: User groups need a paid Slack plan.
  - `name_conflict`: Another Slack user group already uses the handle.
  - `failed`: Better Stack couldn't sync the group with Slack.
  - `no_slack_users`: Nobody on call has connected Slack, so the group tags nobody from this schedule.
  - `reinstall_required`: The Slack integration was connected before Better Stack could manage user groups. Reinstall it to set the group up.
- `error` (String) What's wrong and what to do about it, in the dashboard's words. `null` when status is `active` or `pending`.
- `unconnected_users` (Attributes List) The on-call users the group doesn't include. Present only when you get a single schedule. (see [below for nested schema](#nestedatt--slack_groups--unconnected_users))

<a id="nestedatt--slack_groups--unconnected_users"></a>
### Nested Schema for `slack_groups.unconnected_users`

Read-Only:

- `id` (String) The user's ID as `u123`.
- `first_name` (String) The user's first name, or `null` when it isn't set.
- `last_name` (String) The user's last name, or `null` when it isn't set.
- `email` (String) The user's e-mail address.

<a id="nestedatt--on_call_users"></a>
### Nested Schema for `on_call_users`

Read-Only:

- `id` (String) The person's ID as `u123`.
- `first_name` (String) The person's first name, or `null` when it isn't set.
- `last_name` (String) The person's last name, or `null` when it isn't set.
- `email` (String) The person's e-mail address.
- `phone_numbers` (List of String) The person's phone numbers, such as `+420737123456`, sorted.

<a id="nestedatt--shifts"></a>
### Nested Schema for `shifts`

Read-Only:

- `starts_at` (String (RFC 3339)) When this part of the shift starts. It can be before `from`.
- `ends_at` (String (RFC 3339)) When this part of the shift ends. It can be after `to`.
- `override` (Boolean) Whether an override puts these people on call instead of the rotation.
- `override_id` (String) The override's on-call event ID, which [Delete an on-call event](https://betterstack.com/docs/uptime/api/delete-on-call-calendar-event/) accepts, or `null` for the rotation's own shift.
- `users` (Attributes List) The people on call. (see [below for nested schema](#nestedatt--shifts--users))

<a id="nestedatt--shifts--users"></a>
### Nested Schema for `shifts.users`

Read-Only:

- `id` (String) The person's ID as `u123`.
- `first_name` (String) The person's first name, or `null` when it isn't set.
- `last_name` (String) The person's last name, or `null` when it isn't set.
- `email` (String) The person's e-mail address.
- `phone_numbers` (List of String) The person's phone numbers, sorted.
- `on_holiday` (Boolean) Whether a holiday of theirs covers part of this shift, or of what is left of it while it runs. Escalations skip someone on holiday.
- `holiday_ends_at` (String (RFC 3339)) When that holiday ends, or `null`.
