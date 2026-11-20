---
page_title: "betterstack_monitor_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group that organizes related monitors together.
---

# betterstack_monitor_group (Resource)

A group that organizes related monitors together.

-> Replaces `betteruptime_monitor_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_monitor_group" "example" {
  name = "Backend services"
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `sort_index` (Number) Display order within the team. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `paused` (Boolean) Set to `true` to pause every monitor in the group, or `false` to resume them all. Returned as `true` only when the group has monitors and all of them are paused. Accepted on create, but it has no effect there because a new group has no monitors. Not accepted by the create request; the provider applies it in an immediate follow-up call once the resource exists.

### Read-Only

- `id` (String) The id of this monitor_group.
- `team_name` (String) Name of the team that owns this monitor group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v2/monitor-groups`
- `GET /api/v2/monitor-groups/:id`
- `POST /api/v2/monitor-groups`
- `PATCH /api/v2/monitor-groups/:id`
- `DELETE /api/v2/monitor-groups/:id`

## Import

```shell
terraform import betterstack_monitor_group.example 12345
```
