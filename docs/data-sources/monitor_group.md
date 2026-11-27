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

- `id` (String) The ID of the monitor group to look up.

### Optional

- `team_id` (String) Team that owns this monitor group, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) Group name.
- `sort_index` (Number) The position of the group in your list of monitor groups, lower values first. `null` when no position was set.
- `paused` (Boolean) Whether every monitor in the group is paused, `false` when the group has none.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns this monitor group.
