---
page_title: "betterstack_catalog_record Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog record: one entry in a catalog relation, with values for its attributes.
---

# betterstack_catalog_record (Data Source)

A catalog record: one entry in a catalog relation, with values for its attributes.

## Example Usage

```terraform
data "betterstack_catalog_record" "example" {
  catalog_relation_id = "12345"
  id                  = "12345"
}
```

## Schema

### Address

- `catalog_relation_id` (String, Required) Id of the parent this catalog relation belongs to. Changing it replaces the resource.

### Required

- `id` (String) The id of the catalog_record to look up.

### Read-Only

- `attributes` (Attributes List, see [nested schema](#nestedatt--attributes)) Values for the record's attributes. Each entry sets one catalog attribute, by id or name, to one or more values. Required on create and update. An update replaces all of them.

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

- `attribute` (Attributes, see [nested schema](#nestedatt--attributes--attribute)) The catalog attribute to set, by id or name. Responses include both.
- `values` (Attributes List, see [nested schema](#nestedatt--attributes--values)) Values to set for this attribute. See the note on value shapes.

<a id="nestedatt--attributes--attribute"></a>
### Nested Schema for `attributes.attribute`

- `id` (String)
- `name` (String)

<a id="nestedatt--attributes--values"></a>
### Nested Schema for `attributes.values`

- `type` (String) `String`, or the type of the referenced item.
- `value` (String) The text of a String value.
- `item_id` (String) Id of the referenced item: a number, or a team's t-prefixed id.
- `name` (String) Referenced item's name.
- `email` (String) Referenced user's e-mail.

## Endpoints

- `GET /api/v3/catalog/relations/:catalog_relation_id/records`
- `GET /api/v3/catalog/relations/:catalog_relation_id/records/:id`
- `POST /api/v3/catalog/relations/:catalog_relation_id/records`
- `PATCH /api/v3/catalog/relations/:catalog_relation_id/records/:id`
- `DELETE /api/v3/catalog/relations/:catalog_relation_id/records/:id`
