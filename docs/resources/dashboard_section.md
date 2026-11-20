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

```terraform
resource "betterstack_dashboard_section" "example" {
  dashboard_id = "12345"
  name         = "Key Metrics"
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `dashboard_id` (String, Required) Dashboard this section is on. Can't be changed after the section is created. Changing it replaces the resource.

### Required

- `name` (String) Section name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `y` (Number) Vertical position on the dashboard grid. Defaults to just below the last section or chart. If an update moves the section onto a chart, the layout is tidied up in the background.
- `collapsed` (Boolean) Whether the section starts collapsed. Defaults to `false`.
- `explanation` (String) Explanation of what the section groups, set by the AI tools. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this dashboard_section.

## Endpoints

- `GET /api/v2/dashboards/:dashboard_id/sections`
- `GET /api/v2/dashboards/:dashboard_id/sections/:id`
- `POST /api/v2/dashboards/:dashboard_id/sections`
- `PATCH /api/v2/dashboards/:dashboard_id/sections/:id`
- `DELETE /api/v2/dashboards/:dashboard_id/sections/:id`

## Import

```shell
terraform import betterstack_dashboard_section.example <dashboard_id>/12345
```
