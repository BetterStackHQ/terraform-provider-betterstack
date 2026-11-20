---
page_title: "betterstack_catalog_record Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog record: one entry in a catalog relation, with values for its attributes.
---

# betterstack_catalog_record (Resource)

A catalog record: one entry in a catalog relation, with values for its attributes.

## Example Usage

```terraform
resource "betterstack_catalog_record" "example" {
  catalog_relation_id = "12345"
  attributes          = [{ attribute = { name = "example" } }]
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `catalog_relation_id` (String, Required) Id of the parent this catalog relation belongs to. Changing it replaces the resource.

### Required

- `attributes` (Attributes List, see [nested schema](#nestedatt--attributes)) Values for the record's attributes. Each entry sets one catalog attribute, by id or name, to one or more values. Required on create and update. An update replaces all of them.

### Read-Only

- `id` (String) The id of this catalog_record.

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

- `attribute` (Attributes, see [nested schema](#nestedatt--attributes--attribute)) The catalog attribute to set, by id or name. Responses include both.
- `values` (Attributes List, see [nested schema](#nestedatt--attributes--values)) Values to set for this attribute. See the note on value shapes.

<a id="nestedatt--attributes--attribute"></a>
### Nested Schema for `attributes.attribute`

- `id` (Number)
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

~> The API also serves this object under `/api/v3/catalog/relations/{catalog_relation_id}/records/batch`. One Terraform resource maps to one endpoint family; use `terraform import` or the data source for objects created elsewhere.

-> A value is a plain string, `{type: "String", value: "..."}`, or a reference to an item in the organization. Write a reference as `{type: <ItemType>, item_id: <id>}`, or look the item up with `{type:, name:}` or `{type: "User", email:}`. Responses always return `{type:, item_id:, name:}`, plus email for a User. Reference item types: User, Team, Policy, Schedule, Endpoint, SlackIntegration, LinearIntegration, JiraIntegration, MicrosoftTeamsWebhook, ZapierWebhook, NativeWebhook, PagerDutyWebhook, Source. A team's item_id is its t-prefixed id (t123), in responses and requests.

## Import

```shell
terraform import betterstack_catalog_record.example <catalog_relation_id>/12345
```
