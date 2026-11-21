---
page_title: "betterstack_source_group Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related telemetry sources together.
---

# betterstack_source_group (Resource)

A group that organizes related telemetry sources together.

-> Replaces `logtail_source_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_source_group" "example" {
  name = "Backend sources"
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `sort_index` (Number) Display order among the team's source groups. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this source_group.
- `team_name` (String) Name of the team that owns the group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v1/source-groups`
- `GET /api/v1/source-groups/:id`
- `POST /api/v1/source-groups`
- `PATCH /api/v1/source-groups/:id`
- `DELETE /api/v1/source-groups/:id`

## Import

```shell
terraform import betterstack_source_group.example 12345
```
