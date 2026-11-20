---
page_title: "betterstack_catalog_relation Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog relation: a named table of on-call catalog records and their attributes.
---

# betterstack_catalog_relation (Resource)

A catalog relation: a named table of on-call catalog records and their attributes.

-> Replaces `betteruptime_catalog_relation`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_catalog_relation" "example" {
  name = "Team"
}
```

## Schema

### Required

- `name` (String) Catalog relation name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `organization_id` (String) Organization that owns this catalog relation, as a plain numeric Better Stack ID. Write-only: the API never returns it; Terraform keeps the configured value. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `description` (String) What this catalog relation describes. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `match_mode` (String) How primary attributes combine when matching records: `any` matches a record when any primary attribute matches, `all` when every primary attribute matches. Valid values: `any`, `all`. Defaults to `any`.

### Read-Only

- `id` (String) The id of this catalog_relation.

## Endpoints

- `GET /api/v2/catalog/relations`
- `GET /api/v2/catalog/relations/:id`
- `POST /api/v2/catalog/relations`
- `PATCH /api/v2/catalog/relations/:id`
- `DELETE /api/v2/catalog/relations/:id`

## Import

```shell
terraform import betterstack_catalog_relation.example 12345
```
