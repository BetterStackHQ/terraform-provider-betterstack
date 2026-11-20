---
page_title: "betterstack_catalog_attribute Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog attribute: a named field that every record of a catalog relation can hold.
---

# betterstack_catalog_attribute (Data Source)

A catalog attribute: a named field that every record of a catalog relation can hold.

## Example Usage

```terraform
data "betterstack_catalog_attribute" "example" {
  catalog_relation_id = "12345"
  id                  = "12345"
}
```

## Schema

### Address

- `catalog_relation_id` (String, Required) Id of the parent this catalog relation belongs to. Changing it replaces the resource.

### Required

- `id` (String) The id of the catalog_attribute to look up.

### Read-Only

- `name` (String) Attribute name.
- `primary` (Boolean) Whether this attribute is part of the relation's primary key, used when matching records.
- `position` (Number) Display order of the attribute among the relation's attributes. If omitted, the attribute keeps its place, or gets 0 on create.

## Endpoints

- `GET /api/v2/catalog/relations/:catalog_relation_id/attributes`
- `GET /api/v2/catalog/relations/:catalog_relation_id/attributes/:id`
- `POST /api/v2/catalog/relations/:catalog_relation_id/attributes`
- `PATCH /api/v2/catalog/relations/:catalog_relation_id/attributes/:id`
- `DELETE /api/v2/catalog/relations/:catalog_relation_id/attributes/:id`
