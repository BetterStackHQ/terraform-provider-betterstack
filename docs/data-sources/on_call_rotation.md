---
page_title: "betterstack_on_call_rotation Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An on-call schedule's rotation, the users who are on call in turn and how long each shift lasts.
---

# betterstack_on_call_rotation (Data Source)

An on-call schedule's rotation, the users who are on call in turn and how long each shift lasts.

## Example Usage

```terraform
data "betterstack_on_call_rotation" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the on-call rotation to look up.

### Optional

- `team_id` (String) Team of the rotation's on-call schedule, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `users` (List of String) E-mail addresses of users in the rotation, in shift order. Each must be a responder in the schedule's team.
- `rotation_length` (Number) Number of intervals each shift lasts, a whole number of at least `1`.
- `rotation_interval` (String) Unit for `rotation_length`. Valid values: `hour`, `day`, `week`.
- `start_rotations_at` (String (RFC 3339)) When the rotation schedule starts.
- `end_rotations_at` (String (RFC 3339)) When the rotation schedule ends.
- `timezone` (String) Time zone for `working_hours` and for handovers across daylight saving changes. A new rotation with `working_hours` defaults to UTC.
- `effective_timezone` (String) Time zone that `working_hours` actually use. It's `timezone` when set, else the schedule's default time zone, else `UTC`.
- `apply_from` (String (RFC 3339)) When this change applies, as **Apply changes from** in the web app. Shifts before it are kept. Leave it out to apply the change now.
- `apply_from_timezone` (String) Time zone that `apply_from` is read in when it has no UTC offset. Defaults to UTC.
- `working_hours` (Attributes Set) The windows in which the rotation pages anyone, as up to 50 windows in local time. They are returned as `HH:MM`, sorted by day. A window that ends at or before its start runs overnight, and one with the same start and end covers the whole day. Without it, or with an empty list, the rotation pages around the clock. (see [below for nested schema](#nestedatt--working_hours))

<a id="nestedatt--working_hours"></a>
### Nested Schema for `working_hours`

Read-Only:

- `day` (String) The day of the week the window is on. Valid values: `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`.
- `start_time` (String) When the window starts, as `HH:MM` in 24-hour time.
- `end_time` (String) When the window ends, as `HH:MM` in 24-hour time.
