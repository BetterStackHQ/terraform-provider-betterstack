---
page_title: "betterstack_monitor_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group that organizes related monitors together.
---

# betterstack_monitor_group (Resource)

A group that organizes related monitors together.

-> Replaces `betteruptime_monitor_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a monitor group

```terraform
resource "betterstack_monitor_group" "example" {
  name = "Backend services"
}
```

### Listed first

```terraform
resource "betterstack_monitor_group" "listed_first" {
  name       = "Customer-facing websites"
  sort_index = 0
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

- `sort_index` (Number) The position of the group in your list of monitor groups, lower values first. `null` when no position was set.
- `paused` (Boolean, Computed) Whether every monitor in the group is paused, `false` when the group has none. Set it to `true` to pause them all or `false` to resume them. It has no effect on create.
- `team_id` (String) Team that owns this monitor group, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this monitor group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns this monitor group.

## Import

Import an existing monitor group by its ID. Get the ID via [List monitor groups](https://betterstack.com/docs/uptime/api/list-all-existing-monitor-groups/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_monitor_group.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_monitor_group.example 12345
```
