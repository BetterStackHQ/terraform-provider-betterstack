---
page_title: "betterstack_exploration_group Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related explorations together.
---

# betterstack_exploration_group (Resource)

A group that organizes related explorations together.

-> Replaces `logtail_exploration_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_exploration_group" "example" {
  name = "Production Monitoring"
}
```

## Schema

### Required

- `name` (String) Group name.

### Read-Only

- `id` (String) The id of this exploration_group.
- `team_name` (String) Name of the team that owns the group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v2/exploration-groups`
- `GET /api/v2/exploration-groups/:id`
- `POST /api/v2/exploration-groups`
- `PATCH /api/v2/exploration-groups/:id`
- `DELETE /api/v2/exploration-groups/:id`

-> With an OAuth token, the user needs the dashboards permission, which also covers Explore. A role can also deny one group: the list then doesn't include it, and getting, updating or deleting it, or listing its explorations, returns 403.

-> Served on telemetry.betterstack.com only. Every action returns 404 on rum.betterstack.com.

## Import

```shell
terraform import betterstack_exploration_group.example 12345
```
