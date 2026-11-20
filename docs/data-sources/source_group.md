---
page_title: "betterstack_source_group Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related telemetry sources together.
---

# betterstack_source_group (Data Source)

A group that organizes related telemetry sources together.

## Example Usage

```terraform
data "betterstack_source_group" "example" {
  name = "Backend sources"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Group name.

### Read-Only

- `sort_index` (Number) Display order among the team's source groups.
- `team_name` (String) Name of the team that owns the group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v1/source-groups`
- `GET /api/v1/source-groups/:id`
- `POST /api/v1/source-groups`
- `PATCH /api/v1/source-groups/:id`
- `DELETE /api/v1/source-groups/:id`
