---
page_title: "betterstack_heartbeat_group Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group that organizes related heartbeats together.
---

# betterstack_heartbeat_group (Data Source)

A group that organizes related heartbeats together.

## Example Usage

```terraform
data "betterstack_heartbeat_group" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the heartbeat_group to look up.

### Read-Only

- `team_name` (String) Name of the team that owns this heartbeat group.
- `name` (String) Group name.
- `sort_index` (Number) Display order within the team.
- `paused` (Boolean) Set to `true` to pause every heartbeat in the group, or `false` to resume them all. Returned as `true` only when the group has heartbeats and all of them are paused. Accepted on create, but it has no effect there because a new group has no heartbeats.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v2/heartbeat-groups`
- `GET /api/v2/heartbeat-groups/:id`
- `POST /api/v2/heartbeat-groups`
- `PATCH /api/v2/heartbeat-groups/:id`
- `DELETE /api/v2/heartbeat-groups/:id`
