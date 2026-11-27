---
page_title: "betterstack_source_group Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related telemetry sources together.
---

# betterstack_source_group (Resource)

A group that organizes related telemetry sources together.

-> Replaces `logtail_source_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a source group

```terraform
resource "betterstack_source_group" "example" {
  name = "Backend sources"
}
```

### Set the display order

```terraform
resource "betterstack_source_group" "set_the_display_order" {
  name       = "Frontend sources"
  sort_index = 2
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

- `sort_index` (Number) Display order among the team's source groups. Lower values come first.
- `team_id` (String) Team that owns the group, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this source group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns the group.

## Import

Import an existing source group by its ID. Get the ID from the [`betterstack_source_group` data source](../data-sources/source_group.md), which finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_source_group.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_source_group.example 12345
```
