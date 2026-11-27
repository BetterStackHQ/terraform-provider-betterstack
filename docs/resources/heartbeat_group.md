---
page_title: "betterstack_heartbeat_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group that organizes related heartbeats together.
---

# betterstack_heartbeat_group (Resource)

A group that organizes related heartbeats together.

-> Replaces `betteruptime_heartbeat_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a heartbeat group

```terraform
resource "betterstack_heartbeat_group" "example" {
  name = "Backend services"
}
```

### Listed first

```terraform
resource "betterstack_heartbeat_group" "listed_first" {
  name       = "Database backups"
  sort_index = 0
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

- `sort_index` (Number) The position of the group in your list of heartbeat groups, lower values first. `null` when no position was set.
- `paused` (Boolean, Computed) Whether every heartbeat in the group is paused, `false` when the group has none. Set it to `true` to pause them all or `false` to resume them. Accepted on create, but it has no effect there.
- `team_id` (String) Team that owns this heartbeat group, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this heartbeat group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns this heartbeat group.

## Import

Import an existing heartbeat group by its ID. Get the ID via [List heartbeat groups](https://betterstack.com/docs/uptime/api/list-all-existing-heartbeat-groups/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_heartbeat_group.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_heartbeat_group.example 12345
```
