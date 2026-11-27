---
page_title: "betterstack_catalog_relation Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog relation, a named table of on-call catalog records and their attributes.
---

# betterstack_catalog_relation (Resource)

A catalog relation, a named table of on-call catalog records and their attributes.

-> Replaces `betteruptime_catalog_relation`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Track who owns each service

```terraform
resource "betterstack_catalog_relation" "track_who_owns_each_service" {
  name       = "Services"
  attributes = [{ name = "Service", primary = true }, { name = "Owner" }]
}
```

### Match alerts by service and environment together

```terraform
resource "betterstack_catalog_relation" "match_alerts_by_service_and_environment_together" {
  name        = "Service environments"
  description = "Who owns each service in each environment"
  match_mode  = "all"
  attributes  = [{ name = "Service", primary = true }, { name = "Environment", primary = true }, { name = "Owner" }, { name = "Escalation policy" }]
}
```

## Schema

### Required

- `name` (String) Catalog relation name.
- `attributes` (Attributes List) The relation's attributes, in the order the web app shows them as columns. List primary attributes first, since primary attributes always come before the others. Required on create, and at least one must be primary. An update replaces all attributes with the list, and attributes it leaves out are removed with their values from every record. An update without the list keeps the attributes as they are. In Terraform, an item without an id keeps the id of the existing item with the same `name`, so a reorder or a move updates it in place. An item that matches none keeps the id of the existing item at its position when no other item took that one, so a rename updates it in place too. (see [below for nested schema](#nestedatt--attributes))

### Optional

- `match_mode` (String) How primary attributes combine when matching records. Defaults to `any`.
  - `any`: Matches a record when any primary attribute matches.
  - `all`: Matches a record when every primary attribute matches.
- `organization_id` (String, Forces replacement) Organization that owns this catalog relation, such as `123`. Defaults to the token's own organization. Required when you sign in with OAuth and can access more than one organization. Better Stack never returns it, so Terraform keeps the value you set.
- `description` (String) What this catalog relation describes.

### Read-Only

- `id` (String) The ID of this catalog relation.
- `records_count` (Number) Number of catalog records in this catalog relation.

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

Required:

- `name` (String) Attribute name. Names must be unique within the relation.

Optional:

- `primary` (Boolean) Whether records are matched by this attribute. Defaults to `false`.
- `id` (String) ID of an existing attribute to keep. Leave it out to keep the attribute with the same name, or to add a new one.

Read-Only:

- `position` (Number) Display order, set by the order of the list. Primary attributes count from 0 and added attributes from 1000.

## Import

Import an existing catalog relation by its ID. Get the ID via [List catalog relations](https://betterstack.com/docs/uptime/api/list-all-catalog-relations/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_catalog_relation.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_catalog_relation.example 12345
```
