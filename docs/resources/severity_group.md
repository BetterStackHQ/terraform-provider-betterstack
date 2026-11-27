---
page_title: "betterstack_severity_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A severity group, a set of related severities.
---

# betterstack_severity_group (Resource)

A severity group, a set of related severities.

-> Replaces `betteruptime_severity_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a severity group

```terraform
resource "betterstack_severity_group" "example" {
  name = "Customer-facing"
}
```

### Listed first

```terraform
resource "betterstack_severity_group" "listed_first" {
  name       = "Production outages"
  sort_index = 0
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

- `sort_index` (Number) Display order of the group among the team's severity groups. Lower values come first.
- `team_id` (String) Team that owns this group, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this severity group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns this group.

## Import

Import an existing severity group by its ID. Get the ID via [List severity groups](https://betterstack.com/docs/uptime/api/list-all-existing-severity-groups/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_severity_group.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_severity_group.example 12345
```
