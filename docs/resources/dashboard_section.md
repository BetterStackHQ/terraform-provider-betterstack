---
page_title: "betterstack_dashboard_section Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A full-width section divider that groups charts on a dashboard.
---

# betterstack_dashboard_section (Resource)

A full-width section divider that groups charts on a dashboard.

-> Replaces `logtail_dashboard_section`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a dashboard section

```terraform
resource "betterstack_dashboard_section" "example" {
  dashboard_id = "1234"
  name         = "Errors"
  y            = 10
}
```

### Collapsed below the last chart

```terraform
resource "betterstack_dashboard_section" "collapsed_below_the_last_chart" {
  dashboard_id = "1234"
  name         = "Background jobs"
  collapsed    = true
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `dashboard_id` (String, Required) Dashboard this section is on. Can't be changed after the section is created. Changing it replaces the resource.

### Required

- `name` (String) Section name.

### Optional

- `team_id` (String) Team that owns the section's dashboard, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `y` (Number, Computed) Vertical position on the dashboard grid, from `0` at the top. If an update moves the section onto a chart, the layout is tidied up in the background. Defaults to just below the last section or chart.
- `collapsed` (Boolean) Whether the section starts collapsed. Defaults to `false`.
- `explanation` (String) Explanation of what the section groups, set by the AI tools.

### Read-Only

- `id` (String) The ID of this dashboard section.

## Import

Import an existing dashboard section by its ID as `dashboard_id/id`. Get the ID from the [`betterstack_dashboard_section` data source](../data-sources/dashboard_section.md), which finds it by `name`. `terraform query` with the [`betterstack_dashboard_section` list resource](../list-resources/dashboard_section.md) lists them all, ready to import.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_dashboard_section.example
  id = "<dashboard_id>/12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_dashboard_section.example <dashboard_id>/12345
```
