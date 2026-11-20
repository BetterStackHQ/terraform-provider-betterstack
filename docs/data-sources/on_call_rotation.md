---
page_title: "betterstack_on_call_rotation Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An on-call rotation: the recurring shift configuration for an on-call schedule.
---

# betterstack_on_call_rotation (Data Source)

An on-call rotation: the recurring shift configuration for an on-call schedule.

## Example Usage

```terraform
data "betterstack_on_call_rotation" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the on_call_rotation to look up.

### Read-Only

- `users` (List of String) E-mail addresses of users in the rotation, in shift order. Each must be a responder in the schedule's team.
- `rotation_length` (Number) Number of intervals each shift lasts.
- `rotation_interval` (String) Unit for the rotation length. Valid values: `hour`, `day`, `week`.
- `start_rotations_at` (String (RFC 3339)) When the rotation schedule starts.
- `end_rotations_at` (String (RFC 3339)) When the rotation schedule ends. Must be after start_rotations_at.
- `timezone` (String) Time zone for working hours and for handovers across daylight saving changes. If you leave it out, the previous rotation's time zone is kept. A new rotation with working hours defaults to UTC.
- `effective_timezone` (String) Time zone the working hours actually use.
- `working_hours` (Attributes Set, see [nested schema](#nestedatt--working_hours)) Working hours, as up to 50 windows in local time. They are returned as `HH:MM`, sorted by day. A window that ends at or before its start runs overnight, and one with the same start and end covers the whole day. Leave it out or send `[]` for none.

<a id="nestedatt--working_hours"></a>
### Nested Schema for `working_hours`

- `day` (String) Valid values: `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`.
- `start_time` (String)
- `end_time` (String)

## Endpoints

- `GET /api/v2/on-calls/:id/rotation`
- `POST /api/v2/on-calls/:id/rotation`
- `DELETE /api/v2/on-calls/:id/rotation`
