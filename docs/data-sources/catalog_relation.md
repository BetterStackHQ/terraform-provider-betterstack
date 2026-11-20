---
page_title: "betterstack_catalog_relation Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog relation: a named table of on-call catalog records and their attributes.
---

# betterstack_catalog_relation (Data Source)

A catalog relation: a named table of on-call catalog records and their attributes.

## Example Usage

```terraform
data "betterstack_catalog_relation" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the catalog_relation to look up.

### Read-Only

- `organization_id` (String) Organization that owns this catalog relation, as a plain numeric Better Stack ID.
- `name` (String) Catalog relation name.
- `description` (String) What this catalog relation describes.
- `match_mode` (String) How primary attributes combine when matching records: `any` matches a record when any primary attribute matches, `all` when every primary attribute matches. Valid values: `any`, `all`.

## Endpoints

- `GET /api/v2/catalog/relations`
- `GET /api/v2/catalog/relations/:id`
- `POST /api/v2/catalog/relations`
- `PATCH /api/v2/catalog/relations/:id`
- `DELETE /api/v2/catalog/relations/:id`
