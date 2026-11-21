---
page_title: "betterstack_severity_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A severity group: a set of related severities.
---

# betterstack_severity_group (Resource)

A severity group: a set of related severities.

-> Replaces `betteruptime_severity_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_severity_group" "example" {
  name = "Example"
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `sort_index` (Number) Display order of the group among the team's severity groups. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this severity_group.
- `team_name` (String) Name of the team that owns this group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v3/severity-groups`
- `GET /api/v3/severity-groups/:id`
- `POST /api/v3/severity-groups`
- `PATCH /api/v3/severity-groups/:id`
- `DELETE /api/v3/severity-groups/:id`

## Import

```shell
terraform import betterstack_severity_group.example 12345
```
