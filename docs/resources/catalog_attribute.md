---
page_title: "betterstack_catalog_attribute Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog attribute: a named field that every record of a catalog relation can hold.
---

# betterstack_catalog_attribute (Resource)

A catalog attribute: a named field that every record of a catalog relation can hold.

-> Replaces `betteruptime_catalog_attribute`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_catalog_attribute" "example" {
  catalog_relation_id = "12345"
  name                = "Team"
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `catalog_relation_id` (String, Required) Id of the parent this catalog relation belongs to. Changing it replaces the resource.

### Required

- `name` (String) Attribute name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `primary` (Boolean) Whether this attribute is part of the relation's primary key, used when matching records. Defaults to `false`.
- `position` (Number) Display order of the attribute among the relation's attributes. If omitted, the attribute keeps its place, or gets 0 on create. Defaults to `0`.

### Read-Only

- `id` (String) The id of this catalog_attribute.

## Endpoints

- `GET /api/v2/catalog/relations/:catalog_relation_id/attributes`
- `GET /api/v2/catalog/relations/:catalog_relation_id/attributes/:id`
- `POST /api/v2/catalog/relations/:catalog_relation_id/attributes`
- `PATCH /api/v2/catalog/relations/:catalog_relation_id/attributes/:id`
- `DELETE /api/v2/catalog/relations/:catalog_relation_id/attributes/:id`

## Import

```shell
terraform import betterstack_catalog_attribute.example <catalog_relation_id>/12345
```
