---
page_title: "betterstack_application_group Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  A group that organizes related error-tracking applications together.
---

# betterstack_application_group (Resource)

A group that organizes related error-tracking applications together.

-> Replaces `logtail_errors_application_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create an application group

```terraform
resource "betterstack_application_group" "example" {
  name = "Production apps"
}
```

### Set the display order

```terraform
resource "betterstack_application_group" "set_the_display_order" {
  name       = "Mobile apps"
  sort_index = 2
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

- `sort_index` (Number) Display order among the team's application groups. Lower values come first.
- `team_id` (String) Team that owns the group, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this application group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns the group.

## Import

Import an existing application group by its ID. Get the ID from the [`betterstack_application_group` data source](../data-sources/application_group.md), which finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_application_group.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_application_group.example 12345
```
