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

### Read-Only

- `y` (Number) Vertical position on the dashboard grid. Defaults to just below the last section or chart. If an update moves the section onto a chart, the layout is tidied up in the background.
- `collapsed` (Boolean) Whether the section starts collapsed.
- `explanation` (String) Explanation of what the section groups, set by the AI tools.

## Endpoints

- `GET /api/v2/dashboards/:dashboard_id/sections`
- `GET /api/v2/dashboards/:dashboard_id/sections/:id`
- `POST /api/v2/dashboards/:dashboard_id/sections`
- `PATCH /api/v2/dashboards/:dashboard_id/sections/:id`
- `DELETE /api/v2/dashboards/:dashboard_id/sections/:id`
