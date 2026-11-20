---
page_title: "betterstack_dashboard_group Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related dashboards together.
---

# betterstack_dashboard_group (Resource)

A group that organizes related dashboards together.

-> Replaces `logtail_dashboard_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_dashboard_group" "example" {
  name = "Production Monitoring"
}
```

## Schema

### Required

- `name` (String) Group name.

### Read-Only

- `id` (String) The id of this dashboard_group.
- `team_name` (String) Name of the team that owns the group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v2/dashboard-groups`
- `GET /api/v2/dashboard-groups/:id`
- `POST /api/v2/dashboard-groups`
- `PATCH /api/v2/dashboard-groups/:id`
- `DELETE /api/v2/dashboard-groups/:id`

## Import

```shell
terraform import betterstack_dashboard_group.example 12345
```
