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

- `id` (String) The ID of the heartbeat group to look up.

### Optional

- `team_id` (String) Team that owns this heartbeat group, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) Group name.
- `sort_index` (Number) The position of the group in your list of heartbeat groups, lower values first. `null` when no position was set.
- `paused` (Boolean) Whether every heartbeat in the group is paused, `false` when the group has none.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns this heartbeat group.
