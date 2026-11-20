---
page_title: "betterstack_monitor_group Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group that organizes related monitors together.
---

# betterstack_monitor_group (Data Source)

A group that organizes related monitors together.

## Example Usage

```terraform
data "betterstack_monitor_group" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the monitor_group to look up.

### Read-Only

- `team_name` (String) Name of the team that owns this monitor group.
- `name` (String) Group name.
- `sort_index` (Number) Display order within the team.
- `paused` (Boolean) Set to `true` to pause every monitor in the group, or `false` to resume them all. Returned as `true` only when the group has monitors and all of them are paused. Accepted on create, but it has no effect there because a new group has no monitors.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v2/monitor-groups`
- `GET /api/v2/monitor-groups/:id`
- `POST /api/v2/monitor-groups`
- `PATCH /api/v2/monitor-groups/:id`
- `DELETE /api/v2/monitor-groups/:id`
