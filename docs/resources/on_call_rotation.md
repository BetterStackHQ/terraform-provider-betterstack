---
page_title: "betterstack_on_call_rotation Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An on-call schedule's rotation, the users who are on call in turn and how long each shift lasts.
---

# betterstack_on_call_rotation (Resource)

An on-call schedule's rotation, the users who are on call in turn and how long each shift lasts.

## Example Usage

### Hand over daily, paging only in Prague business hours

```terraform
resource "betterstack_on_call_rotation" "hand_over_daily_paging_only_in_prague_business_hours" {
  id                 = "default"
  users              = ["bob@example.com", "alice@example.com"]
  rotation_length    = 1
  rotation_interval  = "day"
  start_rotations_at = "2026-02-02T00:00:00Z"
  end_rotations_at   = "2028-02-02T00:00:00Z"
  timezone           = "Europe/Prague"
  working_hours      = [{ day = "monday", start_time = "09:00", end_time = "17:00" }, { day = "tuesday", start_time = "09:00", end_time = "17:00" }, { day = "wednesday", start_time = "09:00", end_time = "17:00" }, { day = "thursday", start_time = "09:00", end_time = "17:00" }, { day = "friday", start_time = "09:00", end_time = "17:00" }]
}
```

### Hand over weekly, paging around the clock

```terraform
resource "betterstack_on_call_rotation" "hand_over_weekly_paging_around_the_clock" {
  id                 = "default"
  users              = ["alice@example.com", "bob@example.com", "carol@example.com"]
  rotation_length    = 1
  rotation_interval  = "week"
  start_rotations_at = "2026-12-07T14:00:00Z"
  timezone           = "America/New_York"
}
```

### Switch to 12-hour shifts from a later date

```terraform
resource "betterstack_on_call_rotation" "switch_to_12_hour_shifts_from_a_later_date" {
  id                 = "default"
  users              = ["alice@example.com", "bob@example.com"]
  rotation_length    = 12
  rotation_interval  = "hour"
  start_rotations_at = "2026-12-07T08:00:00Z"
  apply_from         = "2026-12-14T08:00:00Z"
}
```

## Schema

### Required

- `id` (String) Existing object identity. Changing it replaces this Terraform resource.
- `users` (List of String) E-mail addresses of users in the rotation, in shift order. Each must be a responder in the schedule's team.
- `rotation_length` (Number) Number of intervals each shift lasts, a whole number of at least `1`.
- `rotation_interval` (String) Unit for `rotation_length`. Valid values: `hour`, `day`, `week`.
- `start_rotations_at` (String (RFC 3339)) When the rotation schedule starts.

### Optional

- `end_rotations_at` (String (RFC 3339), Computed) When the rotation schedule ends. Must be after `start_rotations_at`. Leave it out to keep the rotation running.
- `team_id` (String) Team of the rotation's on-call schedule, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `timezone` (String, Computed) Time zone for `working_hours` and for handovers across daylight saving changes. A new rotation with `working_hours` defaults to UTC. Leave it out to keep the previous time zone.
- `apply_from` (String (RFC 3339), Computed) When this change applies, as **Apply changes from** in the web app. Shifts before it are kept. Leave it out to apply the change now. Better Stack never returns it, so Terraform keeps the value you set.
- `apply_from_timezone` (String, Computed) Time zone that `apply_from` is read in when it has no UTC offset. Defaults to UTC. Better Stack never returns it, so Terraform keeps the value you set.
- `working_hours` (Attributes Set, Computed) The windows in which the rotation pages anyone, as up to 50 windows in local time. They are returned as `HH:MM`, sorted by day. A window that ends at or before its start runs overnight, and one with the same start and end covers the whole day. Without it, or with an empty list, the rotation pages around the clock. (see [below for nested schema](#nestedatt--working_hours))

### Read-Only

- `effective_timezone` (String) Time zone that `working_hours` actually use. It's `timezone` when set, else the schedule's default time zone, else `UTC`.

<a id="nestedatt--working_hours"></a>
### Nested Schema for `working_hours`

Required:

- `day` (String) The day of the week the window is on. Valid values: `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`.
- `start_time` (String) When the window starts, as `HH:MM` in 24-hour time.
- `end_time` (String) When the window ends, as `HH:MM` in 24-hour time.

-> With `default` as the schedule ID, a global or OAuth token must select exactly one team.

## Import

Import an existing on-call rotation by its ID. Get the ID via [List on-call schedules](https://betterstack.com/docs/uptime/api/list-all-existing-on-call-calendars/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_on_call_rotation.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_on_call_rotation.example 12345
```
