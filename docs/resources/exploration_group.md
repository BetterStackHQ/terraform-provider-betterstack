---
page_title: "betterstack_exploration_group Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related explorations together.
---

# betterstack_exploration_group (Resource)

A group that organizes related explorations together.

-> Replaces `logtail_exploration_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_exploration_group" "example" {
  name = "Production Monitoring"
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

- `team_id` (String) Team that owns the group, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this exploration group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns the group.

## Import

Import an existing exploration group by its ID. Get the ID from the [`betterstack_exploration_group` data source](../data-sources/exploration_group.md), which finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_exploration_group.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_exploration_group.example 12345
```
