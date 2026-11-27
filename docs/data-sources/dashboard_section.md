---
page_title: "betterstack_dashboard_section Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A full-width section divider that groups charts on a dashboard.
---

# betterstack_dashboard_section (Data Source)

A full-width section divider that groups charts on a dashboard.

## Example Usage

```terraform
data "betterstack_dashboard_section" "example" {
  dashboard_id = "12345"
  name         = "Key Metrics"
}
```

## Schema

### Address

- `dashboard_id` (String, Required) Dashboard this section is on. Can't be changed after the section is created. Changing it replaces the resource.

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Section name.

### Optional

- `team_id` (String) Team that owns the section's dashboard, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `y` (Number) Vertical position on the dashboard grid, from `0` at the top. If an update moves the section onto a chart, the layout is tidied up in the background.
- `collapsed` (Boolean) Whether the section starts collapsed.
- `explanation` (String) Explanation of what the section groups, set by the AI tools.
