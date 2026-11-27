---
page_title: "betterstack_catalog_relation Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog relation, a named table of on-call catalog records and their attributes.
---

# betterstack_catalog_relation (Data Source)

A catalog relation, a named table of on-call catalog records and their attributes.

## Example Usage

```terraform
data "betterstack_catalog_relation" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the catalog relation to look up.

### Read-Only

- `organization_id` (String) Organization that owns this catalog relation, such as `123`. Defaults to the token's own organization. Required when you sign in with OAuth and can access more than one organization.
- `name` (String) Catalog relation name.
- `description` (String) What this catalog relation describes.
- `match_mode` (String) How primary attributes combine when matching records.
  - `any`: Matches a record when any primary attribute matches.
  - `all`: Matches a record when every primary attribute matches.
- `records_count` (Number) Number of catalog records in this catalog relation.
- `attributes` (Attributes List) The relation's attributes, in the order the web app shows them as columns. List primary attributes first, since primary attributes always come before the others. Required on create, and at least one must be primary. An update replaces all attributes with the list, and attributes it leaves out are removed with their values from every record. An update without the list keeps the attributes as they are. (see [below for nested schema](#nestedatt--attributes))

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

Read-Only:

- `name` (String) Attribute name. Names must be unique within the relation.
- `primary` (Boolean) Whether records are matched by this attribute. Defaults to `false`.
- `id` (String) ID of an existing attribute to keep. Leave it out to keep the attribute with the same name, or to add a new one.
- `position` (Number) Display order, set by the order of the list. Primary attributes count from 0 and added attributes from 1000.
