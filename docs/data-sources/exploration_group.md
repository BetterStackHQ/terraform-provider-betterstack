---
page_title: "betterstack_exploration_group Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related explorations together.
---

# betterstack_exploration_group (Data Source)

A group that organizes related explorations together.

## Example Usage

```terraform
data "betterstack_exploration_group" "example" {
  name = "Production Monitoring"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Group name.

### Read-Only

- `team_name` (String) Name of the team that owns the group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v2/exploration-groups`
- `GET /api/v2/exploration-groups/:id`
- `POST /api/v2/exploration-groups`
- `PATCH /api/v2/exploration-groups/:id`
- `DELETE /api/v2/exploration-groups/:id`
