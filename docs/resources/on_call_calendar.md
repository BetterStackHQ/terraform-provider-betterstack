---
page_title: "betterstack_on_call_calendar Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An on-call calendar: a named schedule of who is on call and when.
---

# betterstack_on_call_calendar (Resource)

An on-call calendar: a named schedule of who is on call and when.

## Example Usage

```terraform
resource "betterstack_on_call_calendar" "example" {
  name = "My New Calendar"
}
```

## Schema

### Required

- `name` (String) Calendar name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `slack_group_handle` (String) Handle of the Slack user group that follows who is on call, without the @. Null when none. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `use_existing_slack_group` (Boolean) Whether slack_group_handle names an existing Slack user group to take over instead of one created for the calendar. Sending `null` or leaving it out keeps the current setting.

### Read-Only

- `id` (String) The id of this on_call_calendar.
- `default_calendar` (Boolean) Whether this is the team's default on-call calendar.
- `team_name` (String) Name of the team that owns the calendar.
- `slack_groups` (Attributes List, see [nested schema](#nestedatt--slack_groups)) How each connected Slack workspace took the handle. status is pending while Better Stack sets the group up, then shows the state of the link, or reinstall_required. error explains a failure. unconnected_users lists the on-call users the group doesn't include, and only appears when getting a single calendar.
- `on_call_users` (Attributes List, see [nested schema](#nestedatt--on_call_users)) Users on call now, with their id, name, e-mail and phone numbers. When getting a single calendar, date picks another time. Each user's entry under relationships also has meta with email, on_holiday and holiday_ends_at.

<a id="nestedatt--slack_groups"></a>
### Nested Schema for `slack_groups`

- `slack_workspace_id` (String)
- `slack_workspace_name` (String)
- `status` (String)
- `error` (String)
- `unconnected_users` (Attributes List, see [nested schema](#nestedatt--slack_groups--unconnected_users))

<a id="nestedatt--slack_groups--unconnected_users"></a>
### Nested Schema for `slack_groups.unconnected_users`

- `id` (String)
- `first_name` (String)
- `last_name` (String)
- `email` (String)

<a id="nestedatt--on_call_users"></a>
### Nested Schema for `on_call_users`

- `id` (String)
- `first_name` (String)
- `last_name` (String)
- `email` (String)
- `phone_numbers` (List of String)

## Endpoints

- `GET /api/v2/on-calls`
- `GET /api/v2/on-calls/:id`
- `POST /api/v2/on-calls`
- `PATCH /api/v2/on-calls/:id`
- `DELETE /api/v2/on-calls/:id`

## Import

```shell
terraform import betterstack_on_call_calendar.example 12345
```
